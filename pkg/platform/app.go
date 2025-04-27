package platform

import (
	"context"
	"time"

	"github.com/gianglt2198/graphql-gateway-go/pkg"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/cache"
	credis "github.com/gianglt2198/graphql-gateway-go/pkg/infra/cache/redis"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/pubsub"
	psnats "github.com/gianglt2198/graphql-gateway-go/pkg/infra/pubsub/nats"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/serdes"
	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
)

type application struct {
	options []fx.Option
}

type Application interface {
	Start(hooks ...fx.Hook)
}

func buildOptions[T any](config *pkg.Config[T], opts ...fx.Option) []fx.Option {
	logger := monitoring.NewLogger(config.Logger, config.Cfg.Name)

	// Base fx options
	baseOpts := []fx.Option{
		// Supply configs
		fx.Supply(config.Cfg),
		fx.Supply(config.Logger),
		fx.Supply(config.Redis),
		fx.Supply(config.Nats),
		fx.Supply(config.Intermediary),

		// Supply Logger
		fx.Supply(logger),

		// Logger for Fx
		fx.WithLogger(logger.Fx),
	}

	// Add Redis
	baseOpts = append(baseOpts, fx.Provide(fx.Annotate(credis.New, fx.As(new(cache.Cache)))))

	serializers := []serdes.Serializer{
		serdes.NewMsgPack(),
	}

	// Add Serdes
	baseOpts = append(baseOpts, fx.Supply(fx.Annotate(serializers, fx.As(new([]serdes.Serializer)))))

	// Add Nats
	baseOpts = append(baseOpts, fx.Provide(fx.Annotate(psnats.New, fx.As(new(pubsub.Broker[nats.Msg]), new(pubsub.Client), new(pubsub.QueueSubscriber)))))

	return append(baseOpts, opts...)
}

func postBuild() []fx.Option {
	return []fx.Option{
		fx.StartTimeout(1 * time.Second),
		fx.StopTimeout(5 * time.Minute),
	}
}

func CreateApplication[T any](config *pkg.Config[T], fxOptions ...fx.Option) Application {
	return &application{
		options: append(buildOptions(config, fxOptions...), postBuild()...),
	}
}

// Start application
func (a *application) Start(hooks ...fx.Hook) {
	opts := append(a.options, fx.Invoke(run))
	if len(hooks) > 0 {
		opts = append(opts, fx.Invoke(runApplicationHook(hooks...)))
	}
	fx.New(opts...).Run()
}

func run(log *monitoring.AppLogger, lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.GetLogger().Info("Platform lifecycle OnStart")
			return nil
		},
	})
}

// Run after main hook
func runApplicationHook(hooks ...fx.Hook) func(lifecycle fx.Lifecycle) {
	return func(lifecycle fx.Lifecycle) {
		for _, hook := range hooks {
			lifecycle.Append(hook)
		}
	}
}
