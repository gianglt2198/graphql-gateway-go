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
	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SchemaRegistry struct {
	httpClient *http.Client
	logger     *monitoring.Logger

	config config.FederationConfig
	sdlMap map[string]string

	updateDatasourceObservers []DataSourceObserver
}

type SchemaRegistryParams struct {
	fx.In

	Logger *monitoring.Logger
	Config config.FederationConfig
}

func NewSchemaRegistry(params SchemaRegistryParams) *SchemaRegistry {
	return &SchemaRegistry{
		logger:     params.Logger,
		httpClient: http.DefaultClient,
		config:     params.Config,
		sdlMap:     make(map[string]string),
	}
}

func (r *SchemaRegistry) Start(ctx context.Context) {
	r.updateSDLs(ctx)

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.updateSDLs(ctx)
		}
	}
}

func (r *SchemaRegistry) Register(updateDatasourceObserver DataSourceObserver) {
	r.updateDatasourceObservers = append(r.updateDatasourceObservers, updateDatasourceObserver)
}

func (r *SchemaRegistry) updateSDLs(ctx context.Context) {
	r.sdlMap = make(map[string]string)

	var wg sync.WaitGroup
	resultCh := make(chan struct {
		name string
		sdl  string
	})

	for _, subgraph := range r.config.Subgraphs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			sdl, err := r.fetchSchemaSDL(ctx, subgraph.URL)
			if err != nil {
				r.logger.Error("Failed to get sdl.", zap.Error(err))
				return
			}

			select {
			case <-ctx.Done():
			case resultCh <- struct {
				name string
				sdl  string
			}{name: subgraph.Name, sdl: sdl}:
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		r.sdlMap[result.name] = result.sdl
	}

	r.updateObservers()
}

func (r *SchemaRegistry) updateObservers() {
	subgraphsConfig := r.createSubgraphsConfig()

	for i := range r.updateDatasourceObservers {
		r.updateDatasourceObservers[i].UpdateDataSources(subgraphsConfig)
	}
}

func (r *SchemaRegistry) createSubgraphsConfig() []engine.SubgraphConfiguration {
	subgraphConfigs := make([]engine.SubgraphConfiguration, 0, len(r.config.Subgraphs))

	for _, serviceConfig := range r.config.Subgraphs {
		sdl, exists := r.sdlMap[serviceConfig.Name]
		if !exists {
			continue
		}

		subgraphConfig := engine.SubgraphConfiguration{
			Name: serviceConfig.Name,
			URL:  serviceConfig.URL,
			// SubscriptionUrl: serviceConfig.WS,
			SDL: sdl,
		}

		subgraphConfigs = append(subgraphConfigs, subgraphConfig)
	}

	return subgraphConfigs
}

type (
	SchemaResult struct {
		Data struct {
			Service struct {
				SDL string `json:"sdl"`
			} `json:"_service"`
		} `json:"data"`
		Errors []interface{} `json:"errors"`
	}
)

func (r *SchemaRegistry) fetchSchemaSDL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	// Create the GraphQL introspection query
	query := `{
	    _service {
	        sdl
	    }
	}`

	reqBody, _ := json.Marshal(map[string]interface{}{
		"query": query,
	})

	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
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

	var result SchemaResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("error decoding ", err)
		return "", err
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("GrapQL Errors: %v", result.Errors...)
	}

	return result.Data.Service.SDL, nil
}
