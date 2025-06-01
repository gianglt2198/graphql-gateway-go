package db

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
)

type (
	NewClientFn[T any] func(driverName, connection string, opts ...any) *T
)

type DatabaseParams[T any] struct {
	fx.In

	Log         *monitoring.AppLogger
	DBCfg       config.DBConfig
	NewClientFn NewClientFn[T]
}

type DatabaseResult[T any] struct {
	fx.Out

	DBClient *T
}

func NewDBClient[T any](params DatabaseParams[T]) DatabaseResult[T] {
	provider := connect(params.DBCfg, params.NewClientFn)
	return DatabaseResult[T]{DBClient: provider}
}

func connect[T any](cfg config.DBConfig, newFn NewClientFn[T]) *T {
	client := newFn(cfg.Driver, cfg.Connection)

	return client
}
