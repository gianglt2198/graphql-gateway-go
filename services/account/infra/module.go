package infra

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	appConfig "github.com/gianglt2198/federation-go/services/account/config"
)

// Module provides the infrastructure dependencies for the account service
var Module = fx.Module("account-service",
	fx.Provide(NewHTTPServer),
	fx.Invoke(RegisterHTTPServer),
)

// HTTPServer wraps the HTTP server for the account service
type HTTPServer struct {
	cfg    *config.Config[appConfig.AccountConfig]
	server *http.Server
	logger *monitoring.Logger
}

// NewHTTPServer creates a new HTTP server for the account service
func NewHTTPServer(cfg *config.Config[appConfig.AccountConfig], logger *monitoring.Logger) *HTTPServer {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"account"}`))
	})

	// GraphQL endpoint (placeholder for now)
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"service":"account","status":"ready"}}`))
	})

	// GraphQL playground (if enabled)
	mux.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Account Service GraphQL Playground</title>
</head>
<body>
    <h1>Account Service GraphQL Playground</h1>
    <p>GraphQL endpoint: <a href="/graphql">/graphql</a></p>
    <p>Service: Account</p>
</body>
</html>
		`))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: mux,
	}

	return &HTTPServer{
		cfg:    cfg,
		server: server,
		logger: logger,
	}
}

// RegisterHTTPServer registers the HTTP server with the application lifecycle
func RegisterHTTPServer(lc fx.Lifecycle, server *HTTPServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			server.logger.InfoC(ctx, "Starting account service HTTP server",
				zap.String("addr", server.server.Addr))

			go func() {
				if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					server.logger.ErrorC(ctx, "HTTP server error", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.logger.InfoC(ctx, "Stopping account service HTTP server")
			return server.server.Shutdown(ctx)
		},
	})
}
