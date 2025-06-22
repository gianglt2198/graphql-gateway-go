package graphqlservice

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/server"
)

var Module = fx.Module("graphql-server",
	fx.Provide(server.New),
	fx.Invoke(RegisterGraphQLServer),
)

func RegisterGraphQLServer(
	Lifecycle fx.Lifecycle,
	Log *monitoring.Logger,
	GraphQLService server.GraphqlServer,
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
