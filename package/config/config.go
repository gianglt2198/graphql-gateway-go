package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the configuration for all services
type Config[T any] struct {
	App      AppConfig      `mapstructure:"app"`
	Servers  ServerConfig   `mapstructure:"servers"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	NATS     NATSConfig     `mapstructure:"nats"`
	ETCD     ETCDConfig     `mapstructure:"etcd"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Encrypt  EncryptConfig  `mapstructure:"encrypt"`
	Service  T              `mapstructure:"service"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// Load loads configuration from file and environment variables
func Load[T any](configPath string) (*Config[T], error) {
	v := viper.New()

	// Set config file path
	if configPath != "" {
		v.AddConfigPath(configPath)
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Enable environment variable support
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config[T]
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetDatabaseURL returns a formatted database connection URL
func (c *Config[T]) GetDatabaseURL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.Driver,
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetRedisURL returns a formatted Redis connection URL
func (c *Config[T]) GetRedisURL() string {
	if c.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d",
			c.Redis.Password,
			c.Redis.Host,
			c.Redis.Port,
			c.Redis.Database,
		)
	}
	return fmt.Sprintf("redis://%s:%d/%d",
		c.Redis.Host,
		c.Redis.Port,
		c.Redis.Database,
	)
}
