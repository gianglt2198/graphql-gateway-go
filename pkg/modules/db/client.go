package db

import (
	"log"

	"go.uber.org/fx"

	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
)

type (
	NewClientFn[T any] func(driverName, connection string, opts ...any) (*T, error)
)

type DatabaseParams[T any] struct {
	fx.In

	Log         *monitoring.AppLogger
	DBCfg       Config
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

func connect[T any](cfg Config, newFn NewClientFn[T]) *T {
	if !cfg.Enabled {
		return nil
	}

	client, err := newFn(cfg.Driver, cfg.Connection)
	if err != nil {
		log.Fatal("Database Connection Failed!!!", err)
	}

	return client
}
