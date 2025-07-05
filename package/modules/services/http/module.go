package httpservice

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/services/http/server"
)

var Module = fx.Module("http-server",
	fx.Provide(server.New),
	fx.Invoke(RegisterHTTPServer),
)

func RegisterHTTPServer(
	Lifecycle fx.Lifecycle,
	Log *monitoring.Logger,
	HTTPServer server.HTTPServer,
) {
	if HTTPServer == nil {
		return
	}

	Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			Log.Info("Starting HTTP server...")

			go func() {
				if err := HTTPServer.Start(); err != nil {
					Log.Error("Failed to start HTTP server", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			Log.Info("Stopping HTTP server...")

			return HTTPServer.Stop()
		},
	})
}
