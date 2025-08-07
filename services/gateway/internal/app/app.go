package app

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/manager"
	"github.com/gianglt2198/federation-go/services/gateway/graphql"
)

// App represents the gateway application
type App struct {
	FederationManager manager.FederationManager
}

// AppParams defines dependencies for the App
type AppParams struct {
	fx.In

	FederationManager manager.FederationManager
}

// New creates a new gateway application
func New(params AppParams) *App {
	f := graphql.GetAllSchemas()

	// Initialize the federation manager
	params.FederationManager.RegisterSchema(
		"http://gateway.graphql",
		"gateway",
		string(f))
	return &App{
		FederationManager: params.FederationManager,
	}
}

// Run starts the gateway application
func Run(app *App) error {
	return nil
}
