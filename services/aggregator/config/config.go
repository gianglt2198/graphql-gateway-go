package config

import (
	"fmt"

	"github.com/gianglt2198/federation-go/package/config"
)

// AggregatorConfig holds aggregator service specific settings
type AggregatorConfig struct {
	GraphQL GraphQLConfig `mapstructure:"graphql"`
}

// GraphQLConfig holds GraphQL specific settings for aggregator service
type GraphQLConfig struct {
	Playground bool `mapstructure:"playground" default:"true"`
}

// Load loads the configuration for the aggregator service
func Load() (*config.Config[AggregatorConfig], error) {
	baseConfig, err := config.Load[AggregatorConfig]("")
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	return baseConfig, nil
} 