package app

import (
	"context"

	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
)

// App represents the gateway application
type App struct {
	logger *monitoring.Logger
}

// AppParams defines dependencies for the App
type AppParams struct {
	fx.In

	Logger *monitoring.Logger
}

// New creates a new gateway application
func New(params AppParams) *App {
	return &App{
		logger: params.Logger,
	}
}

// Start initializes and starts the gateway application
func (app *App) Start(ctx context.Context) error {
	return nil
}

// Stop gracefully shuts down the gateway application
func (app *App) Stop(ctx context.Context) error {
	app.logger.Info("Shutting down Federation Gateway...")
	return nil
}
