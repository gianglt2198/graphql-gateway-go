package pkg

import (
	"fmt"
	"os"

	env "github.com/caarlos0/env/v11"
	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	credis "github.com/gianglt2198/graphql-gateway-go/pkg/infra/cache/redis"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/etcd"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	psnats "github.com/gianglt2198/graphql-gateway-go/pkg/infra/pubsub/nats"
	"github.com/joho/godotenv"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
)

type Config[T any] struct {
	Cfg          config.Config           `yaml:"common,omitempty"`
	Logger       monitoring.LoggerConfig `yaml:"logger,omitempty"`
	Nats         psnats.Config           `yaml:"nats,omitempty"`
	Intermediary etcd.IntermediaryConfig `yaml:"intermediary,omitempty"`
	Redis        credis.Config           `yaml:"redis,omitempty"`
	App          T                       `yaml:"app,omitempty"`
}

const (
	EnvironmentVariable   = "ENV_FILE"
	DefaultConfigFilePath = "config.yaml"
)

var envLoaded bool

// Load environment variables from .env file by environment.
// Only load once
func LoadEnv() {
	if envLoaded {
		return
	}
	envFile := environmentFile()
	if _, err := os.Stat(envFile); err == nil {
		lo.Must0(godotenv.Load(envFile))
	}
	envLoaded = true
}

func environmentFile() string {
	v := os.Getenv(EnvironmentVariable)
	if v != "" {
		fmt.Printf("Loading env file [%s]", v)
		return v
	}
	return ".env"
}

func LoadConfig[T any](cfg *T) (*T, error) {
	// Load env
	LoadEnv()

	// Load and parse config from environment
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	// Read custom config
	var configFileBytes []byte
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		configFilePath = DefaultConfigFilePath
	}
	configFileBytes, err = os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config in path '%s': %w", configFilePath, err)
	}

	if configFileBytes != nil {
		// Expand environment variables in the config file
		// and unmarshal it into the config struct
		configYamlData := os.ExpandEnv(string(configFileBytes))
		if err := yaml.Unmarshal([]byte(configYamlData), cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	return cfg, nil
}
