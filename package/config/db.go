package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	MigrationsPath  string        `mapstructure:"migrations_path"`
}

func (c *DatabaseConfig) GetURL() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v?search_path=public&sslmode=%v",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}
