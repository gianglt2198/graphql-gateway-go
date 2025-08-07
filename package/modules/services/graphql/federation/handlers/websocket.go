package fhandlers

import (
	"context"
	"net/http"

	fwebsocket "github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/handlers/websocket"
	"github.com/gorilla/websocket"
)

func NewWebSocketFederationMiddleware(ctx context.Context, opts fwebsocket.WebSocketFederationHandlerOptions) func(http.Handler) http.Handler {
	handler := fwebsocket.NewWebSocketFederationHandler(ctx, opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !websocket.IsWebSocketUpgrade(r) {
				next.ServeHTTP(w, r)
				return
			}

			handler.HandleUpgradeRequest(w, r)
		})
	}
}

func NewWebSocketFederationHandler(ctx context.Context, opts fwebsocket.WebSocketFederationHandlerOptions) *fwebsocket.WebSocketFederationHandler {
	return fwebsocket.NewWebSocketFederationHandler(ctx, opts)
}
