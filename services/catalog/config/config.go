package config

import (
	"fmt"

	"github.com/gianglt2198/federation-go/package/config"
)

// CatalogConfig holds catalog service specific settings
type CatalogConfig struct {
	GraphQL GraphQLConfig `mapstructure:"graphql"`
}

// GraphQLConfig holds GraphQL specific settings for catalog service
type GraphQLConfig struct {
	Playground bool `mapstructure:"playground" default:"true"`
}

// Load loads the configuration for the catalog service
func Load() (*config.Config[CatalogConfig], error) {
	baseConfig, err := config.Load[CatalogConfig]("")
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	return baseConfig, nil
} 