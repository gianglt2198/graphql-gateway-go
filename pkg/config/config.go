package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	DEFAULT_CONFIG = "./"
)

type Config[T any] struct {
	Base         BaseConfig         `mapstructure:"common" yaml:"common"`
	DB           DBConfig           `mapstructure:"db" yaml:"db"`
	GQL          GqlServiceConfig   `mapstructure:"gql" yaml:"gql"`
	Redis        RedisConfig        `mapstructure:"redis" yaml:"redis"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring" yaml:"monitoring"`
	S3           S3Config           `mapstructure:"s3" yaml:"s3"`
	Nats         NatsConfig         `mapstructure:"nats" yaml:"nats"`
	Intermediary IntermediaryConfig `mapstructure:"intermediary" yaml:"intermediary"`
	App          T                  `mapstructure:"app" yaml:"app"`
}

type BaseConfig struct {
	Name string `mapstructure:"name" yaml:"name"`
	Env  string `mapstructure:"env" yaml:"env"`
}

func LoadConfig[T any]() *Config[T] {
	v := viper.New()

	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		configFilePath = DEFAULT_CONFIG
	}

	v.AddConfigPath(configFilePath) // optionally look for config in the working directory
	v.SetConfigName("config")       // name of config file (without extension)
	v.SetConfigType("yaml")         // REQUIRED if the config file does not have the extension in the name

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	cfg := &Config[T]{}

	if err := v.Unmarshal(cfg); err != nil {
		panic(err)
	}

	return cfg
}
