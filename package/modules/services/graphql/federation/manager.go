package federation

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/handlers"
	httpServer "github.com/gianglt2198/federation-go/package/modules/services/http/server"
	"github.com/gianglt2198/federation-go/package/modules/services/http/transports"
	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/jensneuse/abstractlogger"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
)

// federationManager implements FederationManager
type federationManager struct {
	appConfig        config.AppConfig
	federationConfig config.FederationConfig

	logger *monitoring.Logger
	mu     sync.RWMutex

	handler        http.Handler
	httpServer     httpServer.HTTPServer
	handlerFactory handlers.HandlerFactory
	registry       *SchemaRegistry
	broker         pubsub.Broker

	// Schema checksum hash
	hash uint64

	readyCh   chan struct{}
	readyOnce *sync.Once
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
		var handlerFactory handlers.HandlerFactoryFn = func(engine *engine.ExecutionEngine) http.Handler {
			return handlers.NewGraphqlHTTPHandler(engine)
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
	handler.ServeHTTP(w, r)
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
		return
	}

	ctx := context.Background()

	var sb strings.Builder
	for _, sub := range subgraphsConfigs {
		sb.WriteString(sub.SDL)
	}
	sdlHashed := utils.Hash(sb.String())
	if f.hash == sdlHashed {
		f.logger.Info("Schema is up-to-date!")
		return
	}
	if f.hash != 0 {
		f.logger.Info("Updating new schema...")
	}
	for i, sub := range subgraphsConfigs {
		subgraphsConfigs[i].SDL = sub.SDL
	}

	transport := transports.NewNatsTransport(transports.NatsTransportParams{
		Upstream: http.DefaultTransport.(*http.Transport),
		Logger:   f.logger,
		Broker:   f.broker,
	})

	client := &http.Client{
		Timeout:   8 * time.Second,
		Transport: transport,
	}

	engineConfigFactory := engine.NewFederationEngineConfigFactory(
		ctx,
		subgraphsConfigs,
		engine.WithFederationHttpClient(client),
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
		f.handler = f.handlerFactory.Make(executionEngine)
	}
	f.mu.Unlock()

	f.readyOnce.Do(func() { close(f.readyCh) })

	if f.hash != 0 {
		f.logger.Info("New schema updated!")
	}
	f.hash = sdlHashed
}
