package main

import (
	"context"
	"log"

	"github.com/gianglt2198/federation-go/package/platform"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/services/catalog/config"
	"github.com/gianglt2198/federation-go/services/catalog/infra"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := platform.NewApp(
		cfg,
		infra.Module,
	)

	app.Run(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
