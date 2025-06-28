package config

import (
	"fmt"
	"os"

	"github.com/gianglt2198/federation-go/package/config"
)

// AccountConfig holds account service specific settings
type AccountConfig struct {
	ExpiredDuration int32 `mapstructure:"expired_duration"` // in hours
}

// Load loads the configuration for the account service
func Load() (*config.Config[AccountConfig], error) {
	configPath := os.Getenv("CONFIG_PATH")
	baseConfig, err := config.Load[AccountConfig](configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	return baseConfig, nil
}
