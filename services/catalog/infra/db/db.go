package db

import (
	"context"

	"github.com/gianglt2198/graphql-gateway-go/catalog/ent"
	"github.com/gianglt2198/graphql-gateway-go/catalog/ent/migrate"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type DBParams struct {
	fx.In
	// Configs
	Db db.Config
	// Common
	Log *monitoring.AppLogger
	// Services
}

type DBResult struct {
	fx.Out

	DbClient *ent.Client
}

func New(params DBParams) DBResult {
	client, err := ent.Open(params.Db.Driver, params.Db.Connection)
	if err != nil {
		params.Log.GetLogger().Fatal("failed opening connection to db", zap.Error(err))
	}
	if err := client.Schema.Create(
		context.Background(),
		migrate.WithGlobalUniqueID(true),
	); err != nil {
		params.Log.GetLogger().Fatal("opening ent client", zap.Error(err))
	}

	return DBResult{
		DbClient: client,
	}
}
