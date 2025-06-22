package infra

import (
	_ "github.com/lib/pq"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
)

var dbModule = fx.Module("db",
	fx.Provide(NewDB),
)

func NewDB(cfg config.DatabaseConfig, logger *monitoring.Logger) *ent.Client {
	db, err := ent.Open(cfg.Driver, cfg.GetURL())
	if err != nil {
		logger.GetLogger().Fatal("failed opening connection", zap.String("driver", cfg.Driver), zap.Error(err))
	}
	return db
}
