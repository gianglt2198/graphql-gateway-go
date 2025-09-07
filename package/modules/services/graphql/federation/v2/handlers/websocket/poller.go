package fwebsocket

import (
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/netpoll"
)

func (h *WebSocketFederationHandler) addConnection(conn net.Conn, handler *WebSocketConnectionHandler) error {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()
	fd := netpoll.SocketFD(conn)
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
				fd := netpoll.SocketFD(conn)
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
