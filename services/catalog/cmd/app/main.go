package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	coreconfig "github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/server/gql"
	"github.com/gianglt2198/graphql-gateway-go/pkg/platform"

	"github.com/gianglt2198/graphql-gateway-go/catalog/config"
	"github.com/gianglt2198/graphql-gateway-go/catalog/graph"
	"github.com/gianglt2198/graphql-gateway-go/catalog/infra/db"
)

func main() {
	cfg := coreconfig.LoadConfig[config.AppConfig]()

	app := platform.CreateApplication(
		cfg,
		db.Module(),
		graph.Module(),
		gql.Module(cfg.Base, cfg.GQL),
	)

	app.Start()
}
