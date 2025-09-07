package redis

import (
	"context"

	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/cache"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

// Module provides NATS pubsub client as an fx module
var Module = []fx.Option{
	fx.Module("cache",
		fx.Provide(
			NewRedis,
			fx.Annotate(
				func(client *Redis) cache.Cache { return client },
				fx.As(new(cache.Cache)),
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, client *Redis, logger *logging.Logger) {
			if client == nil {
				return
			}

			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					logger.GetLogger().Info("Closing NATS connection...")
					return client.Close()
				},
			})
		}),
	),
}
