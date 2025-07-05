package server

import (
	"context"
	runDebug "runtime/debug"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/debug"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/common"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/handlers"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/utils"
	httpServer "github.com/gianglt2198/federation-go/package/modules/services/http/server"
)

type graphqlServer struct {
	appConfig    config.AppConfig
	serverConfig config.GraphQLConfig

	log        *monitoring.Logger
	httpServer httpServer.HTTPServer
	subscriber pubsub.QueueSubscriber

	exec *executor.Executor
}

type ServerParams struct {
	fx.In

	AppConfig    config.AppConfig
	ServerConfig config.GraphQLConfig

	Logger     *monitoring.Logger
	HTTPServer httpServer.HTTPServer
	Subscriber pubsub.QueueSubscriber

	ExecutableSchema graphql.ExecutableSchema
}

func New(params ServerParams) common.GraphqlServer {
	exec := executor.New(params.ExecutableSchema)
	exec.Use(&extension.Introspection{})
	exec.Use(&extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	exec.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	if params.AppConfig.Debug {
		exec.Use(&debug.Tracer{})
	}
	exec.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		return utils.HandleGraphqlError(ctx, err)
	})
	exec.SetRecoverFunc(func(ctx context.Context, err any) error {
		params.Logger.ErrorC(ctx, "error recover", zap.Any("err", err), zap.Any("stack", string(runDebug.Stack())))
		return utils.RecoverFunc(ctx, err)
	})

	if params.ServerConfig.Playground {
		app := params.HTTPServer.GetApp()

		srv := handler.NewDefaultServer(params.ExecutableSchema)

		app.All("/graphql", adaptor.HTTPHandler(srv))

		app.Get("/playground", adaptor.HTTPHandler(playground.ApolloSandboxHandler(
			"Playground",
			"/graphql",
			playground.WithApolloSandboxEndpointIsEditable(true),
			playground.WithApolloSandboxInitialStateIncludeCookies(true),
			playground.WithApolloSandboxInitialStatePollForSchemaUpdates(true),
		)))
	}

	if params.ServerConfig.Enabled {
		if err := handlers.RegisterHandler(params.AppConfig, params.Subscriber, exec); err != nil {
			params.Logger.Fatal("Failed to register graphql handler", zap.Error(err))
		}
	}

	return &graphqlServer{
		appConfig:    params.AppConfig,
		serverConfig: params.ServerConfig,

		log:        params.Logger,
		httpServer: params.HTTPServer,
		subscriber: params.Subscriber,

		exec: exec,
	}
}

func (s *graphqlServer) Start() error {
	s.log.GetLogger().Info("GraphQL service is starting...")

	if s.serverConfig.Playground {
		s.log.GetLogger().Info("GraphQL playground is enabled")
	}

	return nil
}

func (s *graphqlServer) Stop() error {
	s.log.GetLogger().Info("GraphQL service is stopping...")
	return nil
}
