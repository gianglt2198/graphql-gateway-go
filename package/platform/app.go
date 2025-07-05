package platform

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/helpers"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	psnats "github.com/gianglt2198/federation-go/package/infras/pubsub/nats"
	"github.com/gianglt2198/federation-go/package/infras/serdes"
	graphqlservice "github.com/gianglt2198/federation-go/package/modules/services/graphql"
	httpservice "github.com/gianglt2198/federation-go/package/modules/services/http"
)

// App represents the application with all its dependencies
type (
	app struct {
		opts []fx.Option
	}

	App interface {
		Run(...fx.Hook)
	}
)

// NewApp creates a new application with dependency injection
func NewApp[T any](cfg *config.Config[T], modules ...fx.Option) App {
	// Core modules that are always included
	coreModules := []fx.Option{
		// Provide configuration
		fx.Supply(cfg),
		fx.Supply(cfg.App),
		fx.Supply(cfg.Servers.HTTP),
		fx.Supply(cfg.Servers.GraphQL),
		fx.Supply(cfg.Servers.Federation),
		fx.Supply(cfg.Metrics),
		fx.Supply(cfg.Database),
		fx.Supply(cfg.ETCD),
		fx.Supply(cfg.NATS),
		fx.Supply(cfg.Service),
		fx.Supply(cfg.JWT),
		fx.Supply(cfg.Encrypt),
		// Provide logger
		fx.Provide(monitoring.NewLogger),
		// Provide metrics
		fx.Provide(monitoring.NewMetrics),
		// Provide health checker
		fx.Provide(monitoring.NewHealthChecker),
		// Provide JWT helper
		fx.Provide(helpers.NewJWTHelper),
		// Provide encryptor
		fx.Provide(helpers.NewAESCipher),
		// Provide serializers for NATS
		fx.Provide(serdes.NewMsgPack),
		// HTTP server
		httpservice.Module,
		// NATS Server
		psnats.Module,
		// Logger configuration
		fx.WithLogger(func(logger *monitoring.Logger) fxevent.Logger {
			return logger.Fx()
		}),
	}

	if cfg.Servers.GraphQL.Enabled {
		coreModules = append(coreModules, graphqlservice.Module)
	}

	if cfg.Servers.Federation.Enabled {
		coreModules = append(coreModules, graphqlservice.FModule)
	}

	// Combine core modules with provided modules
	allModules := append(coreModules, modules...)

	return &app{
		opts: allModules,
	}
}

// Run starts the application and blocks until it receives a shutdown signal
func (a *app) Run(hooks ...fx.Hook) {
	opts := append(a.opts, fx.Invoke(run))
	if len(hooks) > 0 {
		opts = append(opts, fx.Invoke(runWithHooks(hooks...)))
	}
	fx.New(opts...).Run()
}

func run(lifecycle fx.Lifecycle, log *monitoring.Logger) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Platform lifecycle OnStart")
			return nil
		},
	})
}

// Run after main hook
func runWithHooks(hooks ...fx.Hook) func(lifecycle fx.Lifecycle) {
	return func(lifecycle fx.Lifecycle) {
		for _, hook := range hooks {
			lifecycle.Append(hook)
		}
	}
}
