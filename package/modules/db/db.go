package db

import (
	"context"
	"database/sql"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

func NewDB(cfg config.DatabaseConfig, logger *logging.Logger) *sql.DB {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetURL())
	if err != nil {
		logger.GetLogger().Fatal("failed opening connection", zap.String("driver", cfg.Driver), zap.Error(err))
	}
	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer(
		//otelpgx.WithIncludeQueryParameters(),
		otelpgx.WithTrimSQLInSpanName(),
	)
	pool, err := pgxpool.NewWithConfig(context.TODO(), poolConfig)
	if err != nil {
		logger.GetLogger().Fatal("failed opening connection", zap.String("driver", cfg.Driver), zap.Error(err))
	}
	if err := otelpgx.RecordStats(pool); err != nil {
		logger.GetLogger().Fatal("failed opening connection", zap.String("driver", cfg.Driver), zap.Error(err))
	}
	return stdlib.OpenDBFromPool(pool)
}
