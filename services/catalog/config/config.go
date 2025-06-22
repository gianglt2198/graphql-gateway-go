package config

import (
	"fmt"
	"os"

	"github.com/gianglt2198/federation-go/package/config"
)

// CatalogConfig holds catalog service specific settings
type CatalogConfig struct{}

// Load loads the configuration for the catalog service
func Load() (*config.Config[CatalogConfig], error) {
	configPath := os.Getenv("CONFIG_PATH")

	baseConfig, err := config.Load[CatalogConfig](configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	return baseConfig, nil
}
