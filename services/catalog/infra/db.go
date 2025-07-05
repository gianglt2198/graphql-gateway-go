package infra

import (
	"entgo.io/ent/dialect/sql"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/modules/db"

	"github.com/gianglt2198/federation-go/services/catalog/generated/ent"
)

var dbModule = fx.Module("db",
	fx.Provide(NewDB),
)

func NewDB(cfg config.DatabaseConfig, logger *monitoring.Logger) *ent.Client {
	opts := []ent.Option{
		ent.Driver(sql.OpenDB(cfg.Driver, db.NewDB(cfg, logger))),
	}

	if cfg.Debug {
		opts = append(opts, ent.Debug())
	}
	return ent.NewClient(opts...)
}
