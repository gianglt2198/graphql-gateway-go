package federation

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
	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SchemaRegistry struct {
	httpClient   *http.Client
	logger       *monitoring.Logger
	brokerClient pubsub.Broker

	config config.FederationConfig
	sdlMap map[string]ServiceConfig

	updateDatasourceObservers []DataSourceObserver
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
		sdlMap:       make(map[string]ServiceConfig),
	}
}

func (r *SchemaRegistry) Start(ctx context.Context) {
	r.updateSDLs(ctx)
}

func (r *SchemaRegistry) Register(updateDatasourceObserver DataSourceObserver) {
	r.updateDatasourceObservers = append(r.updateDatasourceObservers, updateDatasourceObserver)
}

func (r *SchemaRegistry) updateSDLs(ctx context.Context) {
	r.sdlMap = make(map[string]ServiceConfig)

	var wg sync.WaitGroup
	resultCh := make(chan ServiceConfig)

	for _, subgraph := range r.config.Subgraphs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				sdl string
				err error
			)
			if r.brokerClient != nil {
				sdl, err = r.fetchSchemaSDL(ctx, subgraph.Name)
				if err != nil {
					r.logger.Error("Failed to get sdl.", zap.Error(err))
					return
				}
			} else {
				sdl, err = r.fetchSchemaSDLWithHTTP(ctx, subgraph.URL)
				if err != nil {
					r.logger.Error("Failed to get sdl.", zap.Error(err))
					return
				}
			}

			if sdl == "" {
				return
			}

			hash := utils.Hash(sdl)
			if _, ok := r.sdlMap[subgraph.Name]; ok && r.sdlMap[subgraph.Name].Hash == hash {
				return
			}

			select {
			case <-ctx.Done():
			case resultCh <- ServiceConfig{
				Name: subgraph.Name,
				URL:  fmt.Sprintf("http://%s", subgraph.Name),
				SDL:  sdl,
				Hash: hash,
			}:
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	isChanged := false
	for result := range resultCh {
		r.sdlMap[result.Name] = result
		isChanged = true
	}

	if isChanged {
		r.updateObservers()
	}
}

func (r *SchemaRegistry) updateObservers() {
	subgraphsConfig := r.createSubgraphsConfig()

	for i := range r.updateDatasourceObservers {
		r.updateDatasourceObservers[i].UpdateDataSources(subgraphsConfig)
	}
}

func (r *SchemaRegistry) createSubgraphsConfig() []engine.SubgraphConfiguration {
	subgraphConfigs := make([]engine.SubgraphConfiguration, 0, len(r.sdlMap))

	for _, subgraph := range r.sdlMap {
		subgraphConfig := engine.SubgraphConfiguration{
			Name: subgraph.Name,
			URL:  subgraph.URL,
			SDL:  subgraph.SDL,
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
