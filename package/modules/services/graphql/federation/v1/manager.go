package federation

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/jensneuse/abstractlogger"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	httpServer "github.com/gianglt2198/federation-go/package/modules/services/http/server"
	"github.com/gianglt2198/federation-go/package/modules/services/http/transports"
	"github.com/gianglt2198/federation-go/package/utils"
)

// federationManager implements FederationManager
type federationManager struct {
	appConfig        config.AppConfig
	federationConfig config.FederationConfig

	logger *monitoring.Logger
	mu     sync.RWMutex

	handler        http.Handler
	httpServer     httpServer.HTTPServer
	handlerFactory HandlerFactory
	registry       *SchemaRegistry
	broker         pubsub.Broker

	// Schema checksum hash
	hash uint64

	readyCh   chan struct{}
	readyOnce *sync.Once
}

// FederationManager defines the interface for federation management
type FederationManager interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	UpdateDataSources(subgraphsConfigs []engine.SubgraphConfiguration)
	Ready() <-chan struct{}
}

type FederationManagerParams struct {
	fx.In

	Logger           *monitoring.Logger
	AppConfig        config.AppConfig
	FederationConfig config.FederationConfig
	HTTPServer       httpServer.HTTPServer
	SchemaRegistry   *SchemaRegistry
	Broker           pubsub.Broker
}

// New creates a new federation manager
func New(params FederationManagerParams) *federationManager {
	f := &federationManager{
		logger:           params.Logger,
		httpServer:       params.HTTPServer,
		appConfig:        params.AppConfig,
		federationConfig: params.FederationConfig,
		registry:         params.SchemaRegistry,
		broker:           params.Broker,
		readyCh:          make(chan struct{}),
		readyOnce:        &sync.Once{},
	}

	f.registry.Register(f)
	go f.registry.Start(context.Background())

	if f.federationConfig.Playground {
		var handlerFactory HandlerFactoryFn = func(logger *monitoring.Logger, engine *engine.ExecutionEngine) http.Handler {
			return NewGraphqlHTTPHandler(logger, engine)
		}

		f.handlerFactory = handlerFactory

		<-f.readyCh

		app := f.httpServer.GetApp()

		if f.appConfig.Debug {
			app.Use(pprof.New())
		}
		app.All("/graphql", adaptor.HTTPHandler(f))
		app.Get(
			"/playground",
			adaptor.HTTPHandlerFunc(playground.ApolloSandboxHandler(
				"Playground",
				"/graphql",
				playground.WithApolloSandboxEndpointIsEditable(true),
				playground.WithApolloSandboxInitialStateIncludeCookies(true),
				playground.WithApolloSandboxInitialStatePollForSchemaUpdates(true),
			)),
		)
	}

	return f
}

func (f *federationManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	handler := f.handler
	f.mu.Unlock()

	if handler == nil {
		http.Error(w, "Federation gateway not ready", http.StatusServiceUnavailable)
		return
	}

	handler.ServeHTTP(w, r)
}

func (f *federationManager) Ready() <-chan struct{} {
	return f.readyCh
}

func (f *federationManager) Start() error {
	f.logger.GetLogger().Info("GraphQL service is starting...")

	return nil
}

func (f *federationManager) Stop() error {
	f.logger.GetLogger().Info("GraphQL service is stopping...")

	return nil
}

func (f *federationManager) UpdateDataSources(subgraphsConfigs []engine.SubgraphConfiguration) {
	if len(subgraphsConfigs) == 0 {
		f.logger.Warn("No subgraph configurations provided")
		return
	}

	ctx := context.Background()

	// Calculate schema hash for change detection
	var schemaBuilder strings.Builder
	for _, config := range subgraphsConfigs {
		schemaBuilder.WriteString(config.SDL)
	}
	sdlHashed := utils.Hash(schemaBuilder.String())

	if f.hash == sdlHashed {
		f.logger.Debug("Schema is up-to-date!")
		return
	}

	f.logger.Info("Updating federation schema",
		zap.Int("subgraph_count", len(subgraphsConfigs)),
		zap.Uint64("schema_hash", sdlHashed),
	)

	// Create HTTP client with custom transport for NATS support
	transport := transports.NewNatsTransport(transports.NatsTransportParams{
		Upstream: http.DefaultTransport.(*http.Transport),
		Logger:   f.logger,
		Broker:   f.broker,
	})

	for i, sub := range subgraphsConfigs {
		subgraphsConfigs[i].SDL = f.extractSDL(sub.SDL)
	}

	client := &http.Client{
		Timeout:   8 * time.Second,
		Transport: transport,
	}

	engineConfigFactory := engine.NewFederationEngineConfigFactory(
		ctx,
		subgraphsConfigs,
		engine.WithFederationHttpClient(client),
		engine.WithFederationSubscriptionType(engine.SubscriptionTypeGraphQLTransportWS),
	)

	engineConfig, err := engineConfigFactory.BuildEngineConfiguration()
	if err != nil {
		f.logger.Error("Failed to build engine config: %s", zap.Error(err))
		return
	}

	executionEngine, err := engine.NewExecutionEngine(
		ctx,
		abstractlogger.NewZapLogger(
			f.logger.GetLogger(),
			abstractlogger.Level(f.logger.GetLogger().Level()),
		),
		engineConfig,
		resolve.ResolverOptions{
			MaxConcurrency:               100,
			Debug:                        f.appConfig.Debug,
			PropagateSubgraphErrors:      true,
			AllowedErrorExtensionFields:  []string{"code", "request_id", "stack_trace"},
			SubgraphErrorPropagationMode: resolve.SubgraphErrorPropagationModePassThrough,
		})
	if err != nil {
		f.logger.Error("Failed to create execution engine: %s", zap.Error(err))
		return
	}

	f.mu.Lock()
	if f.handlerFactory != nil {
		f.handler = f.handlerFactory.Make(f.logger, executionEngine)
	}
	f.mu.Unlock()

	f.readyOnce.Do(func() { close(f.readyCh) })

	if f.hash != 0 {
		f.logger.Info("New schema updated!")
	}
	f.hash = sdlHashed
}

func (f *federationManager) extractSDL(subgraphSDL string) string {
	// Extract the SDL from the subgraphSDL
	// schema, err := gqlparser.LoadSchema(&ast.Source{Name: "subgraph", Input: subgraphSDL})
	// if err != nil {
	// 	return ""
	// }

	// for _, def := range schema.Directives {
	// 	fmt.Println(def.Name)
	// }

	return subgraphSDL
}
