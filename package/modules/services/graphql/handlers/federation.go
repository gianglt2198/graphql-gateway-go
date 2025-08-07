package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"go.uber.org/zap"

	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/execution/graphql"
	"github.com/wundergraph/graphql-go-tools/execution/subscription"
	"github.com/wundergraph/graphql-go-tools/execution/subscription/websocket"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
)

const (
	httpHeaderContentType          string = "Content-Type"
	httpContentTypeApplicationJson string = "application/json"
	httpHeaderUpgrade              string = "Upgrade"
)

type (
	GraphQLHTTPHandler struct {
		logger *monitoring.Logger
		engine *engine.ExecutionEngine
	}
	HandlerFactory interface {
		Make(logger *monitoring.Logger, engine *engine.ExecutionEngine) http.Handler
	}
	HandlerFactoryFn func(logger *monitoring.Logger, engine *engine.ExecutionEngine) http.Handler
)

func (h HandlerFactoryFn) Make(logger *monitoring.Logger, engine *engine.ExecutionEngine) http.Handler {
	return h(logger, engine)
}

func NewGraphqlHTTPHandler(
	logger *monitoring.Logger,
	engine *engine.ExecutionEngine,
) http.Handler {
	return &GraphQLHTTPHandler{
		logger: logger,
		engine: engine,
	}
}

func (g *GraphQLHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isUpgrade := g.isWebsocketUpgrade(r)
	if isUpgrade {
		err := g.upgradeWithNewGoroutine(w, r)
		if err != nil {
			g.logger.Error("GraphQLHTTPHandler.ServeHTTP", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	g.httpHandler(w, r)
}

func (g *GraphQLHTTPHandler) httpHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	var gqlRequest graphql.Request
	if err = graphql.UnmarshalHttpRequest(r, &gqlRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	resultWriter := graphql.NewEngineResultWriterFromBuffer(buf)
	if err = g.engine.Execute(r.Context(), &gqlRequest, &resultWriter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add(httpHeaderContentType, httpContentTypeApplicationJson)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(buf.Bytes()); err != nil {
		return
	}
}

func (g *GraphQLHTTPHandler) upgradeWithNewGoroutine(w http.ResponseWriter, r *http.Request) error {
	upgrader := ws.HTTPUpgrader{
		Protocol: func(s string) bool {
			return s == string(websocket.ProtocolGraphQLTransportWS) || s == string(websocket.ProtocolGraphQLWS)
		},
	}
	conn, _, _, err := upgrader.Upgrade(r, w)
	if err != nil {
		return err
	}
	g.handleWebsocket(r.Context(), conn)
	return nil
}

// handleWebsocket will handle the websocket connection.
func (g *GraphQLHTTPHandler) handleWebsocket(connInitReqCtx context.Context, conn net.Conn) {
	done := make(chan bool)
	errChan := make(chan error)

	executorPool := subscription.NewExecutorV2Pool(g.engine, connInitReqCtx)
	go websocket.HandleWithOptions(done, errChan, conn, executorPool, websocket.HandleOptions{})
	select {
	case err := <-errChan:
		fmt.Printf("Error handling websocket: %v\n", err)
		// g.Error("http.GraphQLHTTPHandler.handleWebsocket()", err.Error())
	case <-done:
	}
}

func (g *GraphQLHTTPHandler) isWebsocketUpgrade(r *http.Request) bool {
	for _, header := range r.Header[httpHeaderUpgrade] {
		if header == "websocket" {
			return true
		}
	}
	return false
}
