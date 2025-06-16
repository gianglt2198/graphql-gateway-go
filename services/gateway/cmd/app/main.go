package main

import (
	"log"

	"github.com/gianglt2198/federation-go/package/platform"
	"github.com/gianglt2198/federation-go/services/gateway/config"
	"github.com/gianglt2198/federation-go/services/gateway/infra"
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

	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
