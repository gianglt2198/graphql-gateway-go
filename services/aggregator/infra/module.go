package infra

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/monitoring"

	appConfig "github.com/gianglt2198/federation-go/services/aggregator/config"
)

// Module provides the infrastructure dependencies for the aggregator service
var Module = fx.Module("aggregator-infra",
	fx.Provide(NewHTTPServer),
	fx.Invoke(RegisterHTTPServer),
)

// HTTPServer wraps the HTTP server for the aggregator service
type HTTPServer struct {
	server *http.Server
	logger *monitoring.Logger
}

// NewHTTPServer creates a new HTTP server for the aggregator service
func NewHTTPServer(cfg *config.Config[appConfig.AggregatorConfig], logger *monitoring.Logger) *HTTPServer {
	port := cfg.Server.Port
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"aggregator"}`))
	})
	
	// GraphQL endpoint (placeholder for now)
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"service":"aggregator","status":"ready"}}`))
	})
	
	// GraphQL playground (if enabled)
	mux.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Aggregator Service GraphQL Playground</title>
</head>
<body>
    <h1>Aggregator Service GraphQL Playground</h1>
    <p>GraphQL endpoint: <a href="/graphql">/graphql</a></p>
    <p>Service: Aggregator</p>
</body>
</html>
		`))
	})
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	return &HTTPServer{
		server: server,
		logger: logger,
	}
}

// RegisterHTTPServer registers the HTTP server with the application lifecycle
func RegisterHTTPServer(lc fx.Lifecycle, server *HTTPServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			server.logger.InfoC(ctx, "Starting aggregator service HTTP server", 
				zap.String("addr", server.server.Addr))
			
			go func() {
				if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					server.logger.ErrorC(ctx, "HTTP server error", zap.Error(err))
				}
			}()
			
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.logger.InfoC(ctx, "Stopping aggregator service HTTP server")
			return server.server.Shutdown(ctx)
		},
	})
} 