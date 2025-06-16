package main

import (
	"log"

	"github.com/gianglt2198/federation-go/package/platform"
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

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run catalog service: %v", err)
	}
} 