package types

import (
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/wundergraph/cosmo/composition-go"
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

type InstanceData struct {
	HostName      string
	ListenAddress string
}

type DataSourceObserverV2 interface {
	UpdateDataSources(subgraphsConfigs []*composition.Subgraph)
}

type FederationHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeWS(c *websocket.Conn)
}
