package platform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
)

// App represents the application with all its dependencies
type App struct {
	fx        *fx.App
	startTime time.Time
}

// AppConfig holds application configuration
type AppConfig struct {
	Name        string
	Version     string
	Environment string
}

// NewApp creates a new application with dependency injection
func NewApp[T any](cfg *config.Config[T], modules ...fx.Option) *App {
	startTime := time.Now()

	// Core modules that are always included
	coreModules := []fx.Option{
		// Provide configuration
		fx.Supply(cfg),
		// Provide logger
		fx.Provide(monitoring.NewLogger),
		// Provide metrics
		fx.Provide(monitoring.NewMetrics),
		// Provide health checker
		fx.Provide(monitoring.NewHealthChecker),
		// Provide HTTP server for metrics and health
		fx.Provide(func(cfg *config.Config[T], metrics *monitoring.Metrics, health *monitoring.HealthChecker) *http.Server {
			mux := http.NewServeMux()
			// Health check endpoint
			mux.Handle("/health", health.HTTPHandler())
			mux.Handle("/health/", health.HTTPHandler())
			// Metrics endpoint
			if cfg.Metrics.Enabled {
				mux.Handle(cfg.Metrics.Path, metrics.Handler())
			}
			// Ready endpoint
			mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			})
			// Live endpoint
			mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			})
			server := &http.Server{
				Addr:         fmt.Sprintf(":%d", cfg.Metrics.Port),
				Handler:      mux,
				ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
				WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
				IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
			}

			return server
		}),

		// Lifecycle hooks
		fx.Invoke(func(lc fx.Lifecycle, cfg *config.Config[T], logger *monitoring.Logger, metrics *monitoring.Metrics, health *monitoring.HealthChecker, server *http.Server) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.InfoC(ctx, "Starting application",
						zap.String("name", cfg.App.Name),
						zap.String("version", cfg.App.Version),
						zap.String("environment", cfg.App.Environment),
					)

					if metrics != nil {
						// Set service uptime
						metrics.SetServiceUptime(cfg.App.Name, time.Since(startTime))
					}

					// Start periodic health checks
					if health != nil {
						go health.StartPeriodicChecks(ctx)
					}

					// Start metrics/health server
					if cfg.Metrics.Enabled {
						go func() {
							logger.InfoC(ctx, "Starting metrics server",
								zap.String("addr", server.Addr),
								zap.String("metrics_path", cfg.Metrics.Path),
							)
							if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
								logger.ErrorC(ctx, "Metrics server failed", zap.Error(err))
							}
						}()
					}

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.InfoC(ctx, "Stopping application")

					// Stop metrics/health server
					if err := server.Shutdown(ctx); err != nil {
						logger.ErrorC(ctx, "Failed to shutdown metrics server", zap.Error(err))
					}

					return nil
				},
			})
		}),

		// Logger configuration
		fx.WithLogger(func(logger *monitoring.Logger) fxevent.Logger {
			return logger.Fx()
		}),
	}

	// Combine core modules with provided modules
	allModules := append(coreModules, modules...)

	app := fx.New(allModules...)

	return &App{
		fx:        app,
		startTime: startTime,
	}
}

// Run starts the application and blocks until it receives a shutdown signal
func (a *App) Run() error {
	// Create context that cancels on interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start the application
	if err := a.fx.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Create context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop the application
	if err := a.fx.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}

	return nil
}
