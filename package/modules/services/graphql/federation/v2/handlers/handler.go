package fhandlers

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"

	"github.com/wundergraph/graphql-go-tools/execution/graphql"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/executor"
	fwebsocket "github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/handlers/websocket"
)

const (
	httpHeaderContentType          string = "Content-Type"
	httpContentTypeApplicationJson string = "application/json"
)

type FederationHandler struct {
	log      *logging.Logger
	executor *executor.Executor

	wsHandler *fwebsocket.WebSocketFederationHandler
}

func NewFederationHandler(log *logging.Logger, executor *executor.Executor) *FederationHandler {
	return &FederationHandler{
		log:      log,
		executor: executor,
	}
}

func (h *FederationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
