package federation

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/handler"
	"github.com/gianglt2198/graphql-gateway-go/gateway/internal/registry"

	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
)

type FederationManager struct {
	registry       *registry.SchemaRegistry
	handlerFactory handler.HandlerFactory

	engine    *engine.ExecutionEngine
	subgraphs []engine.SubgraphConfiguration
	hash      uint64

	gqlHandler http.Handler
	mutex      sync.Mutex
}

// NewFederationManager creates a new federation manager
func NewFederationManager(registry *registry.SchemaRegistry, handlerFactory handler.HandlerFactory) *FederationManager {
	return &FederationManager{
		registry:       registry,
		handlerFactory: handlerFactory,
	}
}

// RefreshFederation rebuilds the federation schema when services change
func (f *FederationManager) RefreshFederation() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// logger := abstractLogger.NewZapLogger(zap.New(), abstractLogger.Level(zapLogger.Level()))

	// Get all service endpoints from your registry
	serviceSdls := f.registry.GetAllSDLs()
	endpoints := f.registry.GetServiceEndpoints() // You'll need to add this method

	subgraphs := make([]engine.SubgraphConfiguration, 0)
	// Add each service to the federation engine
	for serviceName, sdl := range serviceSdls {
		serviceURL, ok := endpoints[serviceName]
		if !ok {
			continue // Skip services without endpoints
		}

		// Store service URL for future reference
		subgraphs = append(subgraphs, engine.SubgraphConfiguration{
			Name: serviceName,
			URL:  serviceURL,
			SDL:  sdl,
		})
	}

	f.subgraphs = subgraphs

	if len(subgraphs) == 0 {
		f.hash = 0
		f.engine = nil
		f.gqlHandler = nil
		return nil
	}

	engineConfigFactory := engine.NewFederationEngineConfigFactory(context.Background(), subgraphs, engine.WithFederationHttpClient(http.DefaultClient))

	engineConfig, err := engineConfigFactory.BuildEngineConfiguration()
	if err != nil {
		// g.logger.Error("Failed to build engine config: %v", abstractLogger.Error(err))
		log.Println("Failed to build engine config: ", err)
		return err
	}

	schema := engineConfig.Schema()
	if schema.Hash() == f.hash {
		log.Println("Schema is up-to-date!")
		return nil
	}

	f.hash = schema.Hash()

	f.engine, err = engine.NewExecutionEngine(context.Background(), nil, engineConfig, resolve.ResolverOptions{
		MaxConcurrency:               10,
		Debug:                        false,
		PropagateSubgraphErrors:      true,
		PropagateSubgraphStatusCodes: true,
		SubgraphErrorPropagationMode: resolve.SubgraphErrorPropagationModePassThrough,
	})
	if err != nil {
		return err
	}

	f.gqlHandler = f.handlerFactory.Make(schema, f.engine)

	return nil
}

func (f *FederationManager) GetHandler() http.Handler {
	return f.gqlHandler
}

// // routeRequest routes a GraphQL request to the appropriate service
// func (f *FederationManager) routeRequest(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
// 	// Implement routing logic based on operation and schema
// 	// This is a placeholder for the actual implementation
// 	return next(ctx)
// }
