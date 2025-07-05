package federation

import (
	"time"

	"github.com/wundergraph/graphql-go-tools/execution/engine"
)

type ServiceConfig struct {
	Name     string
	URL      string
	WS       string
	Fallback func(*ServiceConfig) (string, error)
	Hash     uint64
	SDL      string
}

type DatasourceConfig struct {
	Services        []ServiceConfig
	PollingInterval time.Duration
}

type DataSourceObserver interface {
	UpdateDataSources(subgraphsConfigs []engine.SubgraphConfiguration)
}
