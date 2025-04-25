package pkg

import (
	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	credis "github.com/gianglt2198/graphql-gateway-go/pkg/infra/cache/redis"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/etcd"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
	psnats "github.com/gianglt2198/graphql-gateway-go/pkg/infra/pubsub/nats"
)

type Config struct {
	Cfg          config.Config           `yaml:"common,omitempty"`
	Logger       monitoring.LoggerConfig `yaml:"logger,omitempty"`
	Nats         psnats.Config           `yaml:"nats,omitempty"`
	Intermediary etcd.IntermediaryConfig `yaml:"intermediary,omitempty"`
	Redis        credis.Config           `yaml:"redis,omitempty"`
}
