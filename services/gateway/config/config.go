package config

import (
	"fmt"

	"github.com/gianglt2198/federation-go/package/config"
)

// GatewayConfig holds gateway-specific configuration
type GatewayConfig struct {
	GraphQL GraphQLConfig   `mapstructure:"graphql"`
	Gateway GatewaySettings `mapstructure:"gateway"`
}

// GraphQLConfig holds GraphQL-specific configuration
type GraphQLConfig struct {
	Playground      bool   `mapstructure:"playground"`
	Introspection   bool   `mapstructure:"introspection"`
	Path            string `mapstructure:"path"`
	ComplexityLimit int    `mapstructure:"complexity_limit"`
}

// GatewaySettings holds gateway-specific settings
type GatewaySettings struct {
	SubgraphTimeout       int  `mapstructure:"subgraph_timeout"`
	MaxConcurrentRequests int  `mapstructure:"max_concurrent_requests"`
	EnableTracing         bool `mapstructure:"enable_tracing"`
	EnableMetrics         bool `mapstructure:"enable_metrics"`
}

// Load loads the configuration for the catalog service
func Load() (*config.Config[GatewayConfig], error) {
	baseConfig, err := config.Load[GatewayConfig]("")
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	return baseConfig, nil
}
