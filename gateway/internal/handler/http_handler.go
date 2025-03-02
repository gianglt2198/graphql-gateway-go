package handler

import (
	"bytes"
	"net/http"

	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/execution/graphql"
)

const (
	httpHeaderContentType          string = "Content-Type"
	httpContentTypeApplicationJson string = "application/json"
)

type (
	GraphQLHTTHandler struct {
		schema *graphql.Schema
		engine *engine.ExecutionEngine
	}
	HandlerFactory interface {
		Make(schema *graphql.Schema, engine *engine.ExecutionEngine) http.Handler
	}
	HandlerFactoryFn func(schema *graphql.Schema, engine *engine.ExecutionEngine) http.Handler
)

func (h HandlerFactoryFn) Make(schema *graphql.Schema, engine *engine.ExecutionEngine) http.Handler {
	return h(schema, engine)
}

func NewGraphqlHTTPHandler(
	schema *graphql.Schema,
	engine *engine.ExecutionEngine,
) http.Handler {
	return &GraphQLHTTHandler{
		schema: schema,
		engine: engine,
	}
}

func (g *GraphQLHTTHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
