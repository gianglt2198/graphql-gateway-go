package fhandlers

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/wundergraph/graphql-go-tools/execution/graphql"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/executor"
	fwebsocket "github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/handlers/websocket"
)

const (
	httpHeaderContentType          string = "Content-Type"
	httpContentTypeApplicationJson string = "application/json"
	httpHeaderUpgrade              string = "Upgrade"
)

type FederationHandler struct {
	log      *monitoring.Logger
	executor *executor.Executor

	wsHandler *fwebsocket.WebSocketFederationHandler
}

func NewFederationHandler(log *monitoring.Logger, executor *executor.Executor) *FederationHandler {
	return &FederationHandler{
		log:      log,
		executor: executor,
	}
}

func (h *FederationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if h.isWebsocketUpgrade(r) {
	// 	h.handleWebSocketUpgrade(w, r)
	// 	return
	// }

	h.handleRequest(w, r)
}

func (h *FederationHandler) ServeWS(c *websocket.Conn) {
	h.wsHandler = fwebsocket.NewWebSocketFederationHandler(context.Background(), fwebsocket.WebSocketFederationHandlerOptions{
		Logger:       h.log,
		Executor:     h.executor,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	h.wsHandler.HandleWSUpgradeRequest(c)
}

func (h *FederationHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	var gqlRequest graphql.Request
	if err = graphql.UnmarshalHttpRequest(r, &gqlRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	resultWriter := graphql.NewEngineResultWriterFromBuffer(buf)
	if err = h.executor.Execute(r.Context(), &gqlRequest, &resultWriter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add(httpHeaderContentType, httpContentTypeApplicationJson)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(buf.Bytes()); err != nil {
		return
	}
}

// func (h *FederationHandler) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
// 	h.wsHandler = fwebsocket.NewWebSocketFederationHandler(r.Context(), fwebsocket.WebSocketFederationHandlerOptions{
// 		Logger:       h.log,
// 		Executor:     h.executor,
// 		ReadTimeout:  30 * time.Second,
// 		WriteTimeout: 30 * time.Second,
// 	})

// 	h.wsHandler.HandleUpgradeRequest(w, r)
// }

// func (g *FederationHandler) isWebsocketUpgrade(r *http.Request) bool {
// 	for _, header := range r.Header[httpHeaderUpgrade] {
// 		if header == "websocket" {
// 			return true
// 		}
// 	}
// 	return false
// }
