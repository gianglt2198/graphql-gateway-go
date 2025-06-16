package config

import (
	"fmt"

	"github.com/gianglt2198/federation-go/package/config"
)

// AccountConfig holds account service specific settings
type AccountConfig struct {
	GraphQL GraphQLConfig `mapstructure:"graphql"`
}

// GraphQLConfig holds GraphQL specific settings for account service
type GraphQLConfig struct {
	Playground bool `mapstructure:"playground" default:"true"`
}

// Load loads the configuration for the account service
func Load() (*config.Config[AccountConfig], error) {
	baseConfig, err := config.Load[AccountConfig]("")
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	return baseConfig, nil
} 