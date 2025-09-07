package infra

import (
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/db"
	"github.com/gianglt2198/federation-go/package/modules/db/hooks"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
)

var dbModule = fx.Module("db",
	fx.Provide(NewDB),
)

func NewDB(
	cfg config.DatabaseConfig,
	appCfg config.AppConfig,
	logger *logging.Logger,
	publisher pubsub.Publisher,
) *ent.Client {
	opts := []ent.Option{
		ent.Driver(sql.OpenDB(cfg.Driver, db.NewDB(cfg, logger))),
	}

	if cfg.Debug {
		opts = append(opts, ent.Debug())
	}

	client := ent.NewClient(opts...)

	client.Use(hooks.PublishEntityChangeHook(appCfg.Name, publisher, logger))

	return client
}
