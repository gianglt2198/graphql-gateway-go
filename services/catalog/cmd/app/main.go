package main

import (
	_ "github.com/mattn/go-sqlite3"

	coreconfig "github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/modules/server/gql"
	"github.com/gianglt2198/graphql-gateway-go/pkg/platform"

	"github.com/gianglt2198/graphql-gateway-go/catalog/config"
	"github.com/gianglt2198/graphql-gateway-go/catalog/graph"
	"github.com/gianglt2198/graphql-gateway-go/catalog/infra/db"
)

func main() {
	cfg, err := coreconfig.LoadConfig[config.AppConfig]()
	if err != nil {
		panic(err)
	}

	app := platform.CreateApplication(
		cfg,
		db.Module(),
		graph.Module(),
		gql.Module(cfg.Cfg, cfg.Gql),
	)

	app.Start()
}
