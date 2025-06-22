package config

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Duration  int    `mapstructure:"duration"` // in seconds
}
