package main

import (
	"log"

	"github.com/gianglt2198/federation-go/package/platform"
	"github.com/gianglt2198/federation-go/services/account/config"
	"github.com/gianglt2198/federation-go/services/account/infra"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := platform.NewApp(
		cfg,
		infra.Module,
	)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run account service: %v", err)
	}
} 