package infra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"

	gatewayConfig "github.com/gianglt2198/federation-go/services/gateway/config"
)

// Module provides gateway infrastructure dependencies
var Module = fx.Module("gateway-infra",
	fx.Provide(NewHTTPServer),
	fx.Invoke(RegisterHTTPServer),
)

// NewHTTPServer creates the main HTTP server for the gateway
func NewHTTPServer(
	cfg *config.Config[gatewayConfig.GatewayConfig],
	gatewayCfg *gatewayConfig.GatewayConfig,
	logger *monitoring.Logger,
	metrics *monitoring.Metrics,
) *http.Server {
	mux := http.NewServeMux()

	// GraphQL endpoint (placeholder)
	mux.HandleFunc(gatewayCfg.GraphQL.Path, func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Record metrics
		defer func() {
			duration := time.Since(start)
			metrics.RecordHTTPRequest("gateway", r.Method, r.URL.Path, 200, duration)
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {"message": "Gateway GraphQL endpoint - coming soon"}}`))
	})

	// GraphQL Playground (if enabled)
	if gatewayCfg.GraphQL.Playground {
		mux.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Playground</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
    <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
    <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
    <div id="root">
        <style>
            body { margin: 0; font-family: Open Sans, sans-serif; overflow: hidden; }
            #root { height: 100vh; }
        </style>
    </div>
    <script>
        window.addEventListener('load', function (event) {
            GraphQLPlayground.init(document.getElementById('root'), {
                endpoint: '` + gatewayCfg.GraphQL.Path + `'
            })
        })
    </script>
</body>
</html>
			`))
		})
	}

	// Federation schema endpoint (placeholder)
	mux.HandleFunc("/_federation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"sdl": "# Federation schema coming soon"}`))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return server
}

func RegisterHTTPServer(lc fx.Lifecycle, server *http.Server, logger *monitoring.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.InfoC(ctx, "Starting gateway HTTP server",
				zap.String("addr", server.Addr),
			)

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.ErrorC(ctx, "Gateway HTTP server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.InfoC(ctx, "Stopping gateway HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
