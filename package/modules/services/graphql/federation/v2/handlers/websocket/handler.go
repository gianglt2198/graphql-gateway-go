package fwebsocket

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"go.uber.org/zap"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/netpoll"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/executor"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/handlers/wsprotocol"
)

type WebSocketFederationHandlerOptions struct {
	Logger   *logging.Logger
	Executor *executor.Executor

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	EnableNetPoll         bool
	NetPollTimeout        time.Duration
	NetPollConnBufferSize int
}

type WebSocketFederationHandler struct {
	ctx      context.Context
	logger   *logging.Logger
	executor *executor.Executor

	netPoll       netpoll.Poller
	connections   map[int]*WebSocketConnectionHandler
	connectionsMu sync.RWMutex

	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewWebSocketFederationHandler(ctx context.Context, opts WebSocketFederationHandlerOptions) *WebSocketFederationHandler {
	handler := &WebSocketFederationHandler{
		ctx:      ctx,
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

func (h *WebSocketFederationHandler) HandleWSUpgradeRequest(c *websocket.Conn) {
	conn := newWSConnectionWrapper(c, h.readTimeout, h.writeTimeout)
	protocol, err := wsprotocol.NewProtocol(c.Subprotocol(), conn)
	if err != nil {
		_ = c.Close()
		return
	}

	handler := NewWebSocketConnectionHandler(h.ctx, WebSocketConnectionHandlerOptions{
		Logger:   h.logger,
		Executor: h.executor,

		Protocol:   protocol,
		Connection: conn,
	})

	err = handler.Initialize()
	if err != nil {
		h.logger.Error("Failed to initialize WebSocket connection handler", zap.Error(err))
		handler.Close(false)
		return
	}

	// if h.netPoll != nil {
	// 	err = h.addConnection(c, handler)
	// 	if err != nil {
	// 		handler.Close(true)
	// 	}
	// 	return
	// }

	// Handle messages sync when net poller implementation is not available
	h.handleConnectionSync(handler)
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
