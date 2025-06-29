package federation

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/jensneuse/abstractlogger"
	"go.uber.org/fx"
	"go.uber.org/zap"

	httpServer "github.com/gianglt2198/federation-go/package/modules/services/http/server"
	"github.com/wundergraph/graphql-go-tools/execution/engine"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
)

// federationManager implements FederationManager
type FederationManager struct {
	appConfig        config.AppConfig
	federationConfig config.FederationConfig

	logger *monitoring.Logger
	mu     sync.RWMutex

	handler        http.Handler
	httpServer     httpServer.HTTPServer
	handlerFactory HandlerFactory
	registry       *SchemaRegistry

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
}

// New creates a new federation manager
func New(params FederationManagerParams) *FederationManager {
	f := &FederationManager{
		logger:           params.Logger,
		httpServer:       params.HTTPServer,
		appConfig:        params.AppConfig,
		federationConfig: params.FederationConfig,
		registry:         params.SchemaRegistry,
		readyCh:          make(chan struct{}),
		readyOnce:        &sync.Once{},
	}

	f.registry.Register(f)
	go f.registry.Start(context.Background())

	if f.federationConfig.Playground {
		var handlerFactory HandlerFactoryFn = func(engine *engine.ExecutionEngine) http.Handler {
			return NewGraphqlHTTPHandler(engine)
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

func (f *FederationManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	handler := f.handler
	f.mu.Unlock()
	handler.ServeHTTP(w, r)
}

func (f *FederationManager) Start() error {
	f.logger.GetLogger().Info("GraphQL service is starting...")

	return nil
}

func (f *FederationManager) Stop() error {
	return nil
}

func (f *FederationManager) UpdateDataSources(subgraphsConfigs []engine.SubgraphConfiguration) {
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

	engineConfigFactory := engine.NewFederationEngineConfigFactory(
		ctx,
		subgraphsConfigs,
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
