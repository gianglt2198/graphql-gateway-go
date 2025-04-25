package psnats

import "time"

type Config struct {
	Enabled        bool          `yaml:"enabled" envDefault:"false"`
	Name           string        `yaml:"name"`
	Endpoint       string        `yaml:"endpoint" envDefault:"nats://127.0.0.1:4222"`
	BasePath       string        `yaml:"base_path" envDefault:"local"`
	AllowReconnect bool          `yaml:"allow_reconnect" envDefault:"true"`
	MaxReconnects  int           `yaml:"max_reconnects" envDefault:"500"`
	PingInterval   time.Duration `yaml:"ping_interval" envDefault:"10s"`
}
