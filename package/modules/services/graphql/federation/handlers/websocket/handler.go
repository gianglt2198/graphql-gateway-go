package fwebsocket

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/executor"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/wsprotocol"
	"github.com/gobwas/ws"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/netpoll"
	"go.uber.org/zap"
)

type WebSocketFederationHandlerOptions struct {
	Logger   *monitoring.Logger
	Executor *executor.Executor

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	EnableNetPoll         bool
	NetPollTimeout        time.Duration
	NetPollConnBufferSize int
}

type WebSocketFederationHandler struct {
	ctx      context.Context
	logger   *monitoring.Logger
	executor *executor.Executor

	netPoll       netpoll.Poller
	connections   map[int]*WebSocketConnectionHandler
	connectionsMu sync.RWMutex

	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewWebSocketFederationHandler(ctx context.Context, opts WebSocketFederationHandlerOptions) *WebSocketFederationHandler {
	handler := &WebSocketFederationHandler{
		ctx:      context.Background(),
		logger:   opts.Logger,
		executor: opts.Executor,

		readTimeout:  opts.ReadTimeout,
		writeTimeout: opts.WriteTimeout,
	}

	if opts.EnableNetPoll {
		poller, err := netpoll.NewPoller(opts.NetPollConnBufferSize, opts.NetPollTimeout)
		if err == nil {
			opts.Logger.Debug("Net poller is available")

			handler.netPoll = poller
			handler.connections = make(map[int]*WebSocketConnectionHandler)
			go handler.runPoller()
		}
	}

	return handler
}

func (h *WebSocketFederationHandler) HandleUpgradeRequest(w http.ResponseWriter, r *http.Request) {
	var subProtocol string
	upgrader := ws.HTTPUpgrader{
		Timeout: time.Second * 5,
		Protocol: func(s string) bool {
			if wsprotocol.IsSupportedSubprotocol(s) {
				subProtocol = s
				return true
			}
			return false
		},
	}

	c, _, _, err := upgrader.Upgrade(r, w)
	if err != nil {
		_ = c.Close()
		return
	}

	conn := newWSConnectionWrapper(c, h.readTimeout, h.writeTimeout)
	protocol, err := wsprotocol.NewProtocol(subProtocol, conn)
	if err != nil {
		_ = c.Close()
		return
	}

	handler := NewWebSocketConnectionHandler(h.ctx, WebSocketConnectionHandlerOptions{
		Logger:   h.logger,
		Executor: h.executor,

		Request:        r,
		ResponseWriter: w,

		Protocol:   protocol,
		Connection: conn,
	})

	err = handler.Initialize()
	if err != nil {
		h.logger.Error("Failed to initialize WebSocket connection handler", zap.Error(err))
		handler.Close(false)
		return
	}

	if h.netPoll != nil {
		err = h.addConnection(c, handler)
		if err != nil {
			handler.Close(true)
		}
		return
	}

	// Handle messages sync when net poller implementation is not available

	go h.handleConnectionSync(handler)
}

func (h *WebSocketFederationHandler) handleConnectionSync(handler *WebSocketConnectionHandler) {
	serverDone := h.ctx.Done()
	defer handler.Close(true)

	for {
		select {
		case <-serverDone:
			return
		default:
			msg, err := handler.protocol.ReadMessage()
			if err != nil {
				if isReadTimeout(err) {
					continue
				}
				h.logger.Debug("Client closed connection")
				return
			}
			err = h.HandleMessage(handler, msg)
			if err != nil {
				h.logger.Debug("Handling websocket message", zap.Error(err))
				if errors.Is(err, errClientTerminatedConnection) {
					return
				}
			}
		}
	}
}

func (h *WebSocketFederationHandler) HandleMessage(handler *WebSocketConnectionHandler, msg *wsprotocol.Message) (err error) {
	switch msg.Type {
	case wsprotocol.MessageTypeTerminate:
		return errClientTerminatedConnection
	case wsprotocol.MessageTypePing:
		_ = handler.protocol.Pong(msg)
	case wsprotocol.MessageTypePong:
		// "Furthermore, the Pong message may even be sent unsolicited as a unidirectional heartbeat"
		return nil
	case wsprotocol.MessageTypeSubscribe:
		registration, err := handler.registerSubscription(msg)
		if err != nil {
			h.logger.Warn("Handling subscription registration", zap.Error(err))
			return handler.requestError(fmt.Errorf("error registering subscription id: %s", msg.ID))
		}
		handler.executeSubscription(registration)
	case wsprotocol.MessageTypeComplete:
		err = handler.handleComplete(msg)
		if err != nil {
			h.logger.Warn("Handling complete", zap.Error(err))
		}
	default:
		return handler.requestError(fmt.Errorf("unsupported message type %d", msg.Type))
	}
	return nil
}

func (h *WebSocketFederationHandler) addConnection(conn net.Conn, handler *WebSocketConnectionHandler) error {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()
	fd := socketFd(conn)
	if fd == 0 {
		return fmt.Errorf("unable to get socket fd for conn: %d", handler.connectionID)
	}
	h.connections[fd] = handler
	return h.netPoll.Add(conn)
}

func (h *WebSocketFederationHandler) removeConnection(conn net.Conn, handler *WebSocketConnectionHandler, fd int) {
	h.connectionsMu.Lock()
	delete(h.connections, fd)
	h.connectionsMu.Unlock()
	err := h.netPoll.Remove(conn)
	if err != nil {
		h.logger.Warn("Removing connection from net poller", zap.Error(err))
	}
	handler.Close(true)
}

func (h *WebSocketFederationHandler) runPoller() {
	done := h.ctx.Done()
	defer func() {
		h.connectionsMu.Lock()
		_ = h.netPoll.Close(true)
		h.connectionsMu.Unlock()
	}()
	for {
		select {
		case <-done:
			return
		default:
			connections, err := h.netPoll.Wait(128)
			if err != nil {
				h.logger.Warn("Net Poller wait", zap.Error(err))
				continue
			}
			for i := range len(connections) {
				if connections[i] == nil {
					continue
				}
				conn := connections[i].(netpoll.ConnImpl)
				// check if the connection is still valid
				fd := socketFd(conn)
				h.connectionsMu.RLock()
				handler, exists := h.connections[fd]
				h.connectionsMu.RUnlock()

				if !exists {
					h.logger.Debug("Connection not found", zap.Int("fd", fd))
					continue
				}

				if fd == 0 {
					h.logger.Debug("Invalid socket fd", zap.Int("fd", fd))
					h.removeConnection(conn, handler, fd)
					continue
				}

				msg, err := handler.protocol.ReadMessage()
				if err != nil {
					h.logger.Debug("Client closed connection", zap.Error(err))
					h.removeConnection(conn, handler, fd)
					continue
				}
				err = h.HandleMessage(handler, msg)
				if err != nil {
					h.logger.Debug("Handling websocket message", zap.Error(err))
					if errors.Is(err, errClientTerminatedConnection) {
						h.removeConnection(conn, handler, fd)
						continue
					}
				}
			}
		}
	}
}
