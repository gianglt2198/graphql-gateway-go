package graphqlservice

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/common"
	federation "github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v1"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/manager"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/registry"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/server"
)

var Module = fx.Module("graphql-module",
	fx.Provide(fx.Annotate(server.New, fx.As(new(common.GraphqlServer)))),
	fx.Invoke(RegisterGraphQLServer),
)

var FModule = fx.Module("federation-module",
	fx.Provide(federation.NewSchemaRegistry),
	fx.Provide(fx.Annotate(federation.New, fx.As(new(common.GraphqlServer)))),
	fx.Invoke(RegisterGraphQLServer),
)

var FModuleV2 = fx.Module("federation-module-v2",
	fx.Provide(registry.NewSchemaRegistry),
	fx.Provide(manager.New,
		fx.Annotate(
			func(m manager.FederationManager) common.GraphqlServer { return m },
			fx.As(new(common.GraphqlServer))),
	),
	fx.Invoke(RegisterGraphQLServer),
)

func RegisterGraphQLServer(
	Lifecycle fx.Lifecycle,
	Log *logging.Logger,
	GraphQLService common.GraphqlServer,
) {
	Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			Log.Info("Starting GraphQL service...")

			go func() {
				if err := GraphQLService.Start(); err != nil {
					Log.Error("Failed to start GraphQL service", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			Log.Info("Stopping GraphQL service...")

			return GraphQLService.Stop()
		},
	})
}
