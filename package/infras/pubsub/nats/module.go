package psnats

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
)

// Module provides NATS pubsub client as an fx module
var Module = []fx.Option{
	fx.Module("nats",
		fx.Provide(
			NewNATSClient,
			fx.Annotate(
				func(client *natsProvider) pubsub.Client { return client },
				fx.As(new(pubsub.Client)),
			),
			fx.Annotate(
				func(client *natsProvider) pubsub.QueueSubscriber { return client },
				fx.As(new(pubsub.QueueSubscriber)),
			),
			fx.Annotate(
				func(client *natsProvider) pubsub.Broker { return client },
				fx.As(new(pubsub.Broker)),
			),
			fx.Annotate(
				func(client *natsProvider) pubsub.QueueClient { return client },
				fx.As(new(pubsub.QueueClient)),
			),
		),
		fx.Invoke(func(lc fx.Lifecycle, client *natsProvider, logger *logging.Logger) {
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

// NewNATSClient creates a new NATS client with dependency injection support
func NewNATSClient(params NatsParams) *natsProvider {
	if params.Config.Endpoint == "" {
		return nil
	}

	provider := New(params)

	// Add health check
	params.Log.GetLogger().Info("NATS client initialized successfully",
		zap.String("endpoint", params.Config.Endpoint),
		zap.String("name", params.Config.Name),
	)

	return provider
}
