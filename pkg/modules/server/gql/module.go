package gql

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/server/gql/service"
	"github.com/gianglt2198/graphql-gateway-go/pkg/utils/async"
)

func Module(cfg config.Config, scfg service.GqlServiceConfig) fx.Option {
	return fx.Module(cfg.Name, buildServerModuleOptions(scfg)...)
}

func buildServerModuleOptions(cfg service.GqlServiceConfig) []fx.Option {
	opts := []fx.Option{}

	if !cfg.Enabled {
		return opts
	}

	opts = append(opts, fx.Supply(cfg), fx.Provide(service.NewGqlServer), fx.Invoke(startGraphql))
	return opts
}

func startGraphql(log *monitoring.AppLogger, server service.GqlServer, lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.Start(); err != nil {
					log.GetLogger().Error("failed to start server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return async.WaitAll(
				async.Errable(server.Stop),
			)
		},
	})
}
