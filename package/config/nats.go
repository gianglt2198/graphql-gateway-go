package config

import "time"

// NATSConfig holds NATS configuration
type NATSConfig struct {
	Enabled        bool          `mapstructure:"enabled" yaml:"enabled" envDefault:"false"`
	Name           string        `mapstructure:"name" yaml:"name"`
	Endpoint       string        `mapstructure:"endpoint" yaml:"endpoint" envDefault:"nats://127.0.0.1:4222"`
	BasePath       string        `mapstructure:"base_path" yaml:"base_path" envDefault:"local"`
	AllowReconnect bool          `mapstructure:"allow_reconnect" yaml:"allow_reconnect" envDefault:"true"`
	MaxReconnects  int           `mapstructure:"max_reconnects" yaml:"max_reconnects" envDefault:"500"`
	PingInterval   time.Duration `mapstructure:"ping_interval" yaml:"ping_interval" envDefault:"10s"`
}
