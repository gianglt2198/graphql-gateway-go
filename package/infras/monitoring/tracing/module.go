package tracing

import (
	"context"

	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

// Module provides Tracing OTel client as an fx module
var Module = []fx.Option{
	fx.Module("tracing",
		fx.Provide(
			NewTracing,
		),
		fx.Invoke(func(lc fx.Lifecycle, client *tracingClient, logger *logging.Logger) {
			if client == nil {
				return
			}

			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					logger.GetLogger().Info("Closing tracing connection...")
					return client.shutdown(ctx)
				},
			})
		}),
	),
}
