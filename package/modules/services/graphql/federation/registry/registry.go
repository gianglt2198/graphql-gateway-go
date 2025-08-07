package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/types"
	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/wundergraph/cosmo/composition-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SchemaRegistry struct {
	httpClient   *http.Client
	logger       *monitoring.Logger
	brokerClient pubsub.Broker

	config config.FederationConfig
	sdlMap map[string]types.ServiceConfig

	updateDatasourceObservers []types.DataSourceObserverV2
	mu                        sync.RWMutex
}

type SchemaRegistryParams struct {
	fx.In

	Logger       *monitoring.Logger
	Config       config.FederationConfig
	BrokerClient pubsub.Broker
}

type Result struct {
	Data struct {
		Service struct {
			SDL string `json:"sdl"`
		} `json:"_service"`
	} `json:"data"`
	Errors []GQLErr `json:"errors,omitempty"`
}

type GQLErr struct {
	Message string `json:"message"`
}

const ServiceDefinitionQuery = `
	{
		"query": "query __ApolloGetServiceDefinition__ { _service { sdl } }",
		"operationName": "__ApolloGetServiceDefinition__",
		"variables": {}
	}`

func NewSchemaRegistry(params SchemaRegistryParams) *SchemaRegistry {
	return &SchemaRegistry{
		logger:       params.Logger,
		httpClient:   http.DefaultClient,
		brokerClient: params.BrokerClient,
		config:       params.Config,
		sdlMap:       make(map[string]types.ServiceConfig),
	}
}

func (r *SchemaRegistry) Start(ctx context.Context) {
	r.updateSDLs(ctx)
}

func (r *SchemaRegistry) Register(updateDatasourceObserver types.DataSourceObserverV2) {
	r.updateDatasourceObservers = append(r.updateDatasourceObservers, updateDatasourceObserver)
}

func (r *SchemaRegistry) RegisterSchema(url, name, sdl string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sdlMap[name]; exists {
		r.logger.Warn("Schema already registered", zap.String("name", name))
		return
	}

	r.sdlMap[name] = types.ServiceConfig{
		Name: name,
		URL:  url,
		SDL:  sdl,
		Hash: utils.Hash(sdl),
	}
}

func (r *SchemaRegistry) updateSDLs(ctx context.Context) {
	// Convert config subgraphs to ServiceConfig
	var serviceConfigs []types.ServiceConfig
	for _, subgraph := range r.config.Subgraphs {
		svc := types.ServiceConfig{
			Name: subgraph.Name,
			URL:  subgraph.URL,
		}

		if svc.URL == "" {
			svc.URL = fmt.Sprintf("http://%s", svc.Name)
		}

		serviceConfigs = append(serviceConfigs, svc)
	}

	var wg sync.WaitGroup
	resultCh := make(chan types.ServiceConfig)

	for _, serviceConfig := range serviceConfigs {
		wg.Add(1)
		go func(sc types.ServiceConfig) {
			defer wg.Done()
			var (
				sdl string
				err error
			)

			sdl, err = r.fetchSchemaFromService(ctx, sc.Name)
			if err != nil {
				r.logger.Error("Failed to fetch schema",
					zap.String("service", sc.Name),
					zap.String("url", sc.URL),
					zap.Error(err),
				)
				return
			}

			if sdl == "" {
				return
			}

			hash := utils.Hash(sdl)
			if _, ok := r.sdlMap[sc.Name]; ok && r.sdlMap[sc.Name].Hash == hash {
				return
			}

			sc.SDL = sdl
			sc.Hash = utils.Hash(sdl)
			resultCh <- sc

		}(serviceConfig)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results and check for changes
	var updatedConfigs []types.ServiceConfig
	for result := range resultCh {
		r.mu.Lock()
		r.sdlMap[result.Name] = result
		r.mu.Unlock()

		updatedConfigs = append(updatedConfigs, result)
	}

	// Notify observers if there are changes
	if len(updatedConfigs) > 0 {
		r.updateObservers()
	}
}

func (r *SchemaRegistry) updateObservers() {
	subgraphsConfig := r.createSubgraphsConfig()

	for i := range r.updateDatasourceObservers {
		r.updateDatasourceObservers[i].UpdateDataSources(subgraphsConfig)
	}
}

func (r *SchemaRegistry) createSubgraphsConfig() []*composition.Subgraph {
	subgraphConfigs := make([]*composition.Subgraph, 0, len(r.sdlMap))

	for _, subgraph := range r.sdlMap {
		subgraphConfig := &composition.Subgraph{
			Name:   subgraph.Name,
			URL:    subgraph.URL,
			Schema: subgraph.SDL,
		}

		subgraphConfigs = append(subgraphConfigs, subgraphConfig)
	}

	return subgraphConfigs
}

func (r *SchemaRegistry) fetchSchemaSDLWithHTTP(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Body = io.NopCloser(bytes.NewBuffer([]byte(ServiceDefinitionQuery)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("error decoding ", err)
		return "", err
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("GrapQL Errors: %v", result.Errors)
	}

	return result.Data.Service.SDL, nil
}

func (r *SchemaRegistry) fetchSchemaSDL(ctx context.Context, url string) (string, error) {
	result := Result{}
	err := r.brokerClient.Request(ctx, url, []byte(ServiceDefinitionQuery), nil, 5*time.Second, &result)
	if err != nil {
		return "", fmt.Errorf("do request: %v", err)
	}

	if result.Errors != nil {
		return "", fmt.Errorf("response error: %v", result.Errors)
	}
	return result.Data.Service.SDL, nil
}

func (r *SchemaRegistry) fetchSchemaFromService(ctx context.Context, url string) (string, error) {
	// Try NATS-based request first (for internal services)
	if r.brokerClient != nil {
		schema, err := r.fetchSchemaSDL(ctx, url)
		if err == nil {
			return schema, nil
		}
		r.logger.Debug("NATS fetch failed, trying HTTP", zap.Error(err))
	}

	// Fallback to HTTP request
	return r.fetchSchemaSDLWithHTTP(ctx, url)
}
