package service

import (
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/debug"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"

	"github.com/gianglt2198/graphql-gateway-go/pkg/utils"
)

type registerSchema struct {
	GraphqlPath       string
	GraphqlPort       int
	PlaygroundPath    string
	PlaygroundEnabled bool
	Debug             bool
	Schema            graphql.ExecutableSchema
}

func registerWithSchema(cfg registerSchema) error {
	srv := handler.New(cfg.Schema)
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Options{})
	srv.Use(&extension.Introspection{})
	srv.Use(&extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	if cfg.Debug {
		srv.Use(&debug.Tracer{})
	}

	if cfg.PlaygroundEnabled {
		srv.AddTransport(transport.GET{})
		http.Handle(cfg.PlaygroundPath, playground.ApolloSandboxHandler(
			"Playground",
			cfg.GraphqlPath,
			playground.WithApolloSandboxHideCookieToggle(true),
			playground.WithApolloSandboxEndpointIsEditable(true),
			playground.WithApolloSandboxInitialStatePollForSchemaUpdates(true),
		))
	}

	handler := cors.AllowAll().Handler(srv)
	http.Handle(cfg.GraphqlPath, applyRequestIDInContext(handler))

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.GraphqlPort), nil)
}

func applyRequestIDInContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, _ = utils.ApplyRequestIDWithContext(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
