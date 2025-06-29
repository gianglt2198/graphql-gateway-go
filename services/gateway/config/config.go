package config

import (
	"fmt"
	"os"

	"github.com/gianglt2198/federation-go/package/config"
)

// GatewayConfig holds gateway-specific configuration
type GatewayConfig struct{}

// Load loads and validates the configuration for the gateway service
func Load() (*config.Config[GatewayConfig], error) {
	configPath := os.Getenv("CONFIG_PATH")
	baseConfig, err := config.Load[GatewayConfig](configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load gateway config: %w", err)
	}

	return baseConfig, nil
}
