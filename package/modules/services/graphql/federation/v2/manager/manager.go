package manager

import (
	"context"
	"net/http"
	"sync"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/contrib/websocket"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	composition "github.com/wundergraph/cosmo/composition-go"
	nodev1 "github.com/wundergraph/cosmo/router/gen/proto/wg/cosmo/node/v1"
	routerCfg "github.com/wundergraph/cosmo/router/pkg/config"
	"github.com/wundergraph/cosmo/router/pkg/pubsub/datasource"
	"github.com/wundergraph/cosmo/router/pkg/statistics"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/common"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/types"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/executor"
	fhandlers "github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/handlers"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/loader"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/registry"
	httpServer "github.com/gianglt2198/federation-go/package/modules/services/http/server"
)

// federationManager implements FederationManager
type federationManager struct {
	appConfig        config.AppConfig
	federationConfig config.FederationConfig

	logger *logging.Logger
	mu     sync.RWMutex

	handler    types.FederationHandler
	httpServer httpServer.HTTPServer
	registry   *registry.SchemaRegistry
	broker     pubsub.Broker

	schemas []*composition.Subgraph

	pubsubProviders []datasource.Provider

	readyCh   chan struct{}
	readyOnce *sync.Once
}

// FederationManager defines the interface for federation management
type FederationManager interface {
	common.GraphqlServer
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	UpdateDataSources(subgraphsConfigs []*composition.Subgraph)
	Ready() <-chan struct{}
	RegisterSchema(url, name, sdl string)
}

type FederationManagerParams struct {
	fx.In

	Logger           *logging.Logger
	AppConfig        config.AppConfig
	FederationConfig config.FederationConfig
	HTTPServer       httpServer.HTTPServer
	SchemaRegistry   *registry.SchemaRegistry
	Broker           pubsub.Broker
}

// New creates a new federation manager
func New(params FederationManagerParams) FederationManager {
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
		app := f.httpServer.GetApp()

		if f.appConfig.Debug {
			app.Use(pprof.New())
		}

		app.Use("/ws", func(c *fiber.Ctx) error {
			if !websocket.IsWebSocketUpgrade(c) {
				return fiber.ErrUpgradeRequired
			}
			return c.Next()
		})

		app.Get("/ws", websocket.New(f.ServeWS, websocket.Config{
			Subprotocols: []string{"graphql-transport-ws", "graphql-ws"},
		}))

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

func (f *federationManager) RegisterSchema(url, name, sdl string) {
	f.registry.RegisterSchema(url, name, sdl)

	go f.registry.Start(context.Background())
}

func (f *federationManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.handler == nil {
		http.Error(w, "Federation gateway not ready", http.StatusServiceUnavailable)
		return
	}

	f.handler.ServeHTTP(w, r)
}

func (f *federationManager) ServeWS(c *websocket.Conn) {
	if f.handler == nil {
		_ = c.WriteMessage(websocket.CloseMessage, []byte("Federation gateway not ready"))
		return
	}
	f.handler.ServeWS(c)
}

func (f *federationManager) Ready() <-chan struct{} {
	return f.readyCh
}

func (f *federationManager) Start() error {
	f.logger.GetLogger().Info("GraphQL service is starting...")

	<-f.readyCh

	return nil
}

func (f *federationManager) Stop() error {
	f.logger.GetLogger().Info("GraphQL service is stopping...")

	if err := f.shutdownProviders(context.Background()); err != nil {
		f.logger.GetLogger().Error("Failed to shutdown pubsub providers", zap.Error(err))
	}

	return nil
}

func (f *federationManager) UpdateDataSources(subgraphsConfigs []*composition.Subgraph) {
	if len(subgraphsConfigs) == 0 {
		f.logger.Warn("No subgraph configurations provided")
		return
	}

	f.logger.Info("Updating federation schema",
		zap.Int("subgraph_count", len(subgraphsConfigs)),
	)

	if len(f.schemas) > 0 {
		subgraphsConfigs = append(subgraphsConfigs, f.schemas...)
	}

	resultJSON, err := composition.BuildRouterConfiguration(subgraphsConfigs...)
	if err != nil {
		f.logger.Error("Failed to build router configuration", zap.Error(err))
		return
	}

	var routerConfig nodev1.RouterConfig
	if err := protojson.Unmarshal([]byte(resultJSON), &routerConfig); err != nil {
		f.logger.Error("Failed to unmarshal router configuration", zap.Error(err))
		return
	}

	routerEngineConfig := &loader.RouterEngineConfiguration{
		Execution: routerCfg.EngineExecutionConfiguration{},
		Events: routerCfg.EventsConfiguration{
			Providers: routerCfg.EventProviders{
				Nats: []routerCfg.NatsEventSource{
					{
						ID:  "default",
						URL: "nats://localhost:4223",
					},
				},
			},
		},
		SubgraphErrorPropagation: routerCfg.SubgraphErrorPropagationConfiguration{
			Enabled:                 true,
			AllowAllExtensionFields: true,
			// AllowedErrorExtensionFields: []string{"code", "request_id", "stack_trace"},
			Mode:         routerCfg.SubgraphErrorPropagationModePassthrough,
			RewritePaths: true,
		},
	}

	engineStats := statistics.NewNoopEngineStats()

	ecbParams := executor.ExecutorConfigurationBuildParams{
		EngineConfig:       routerConfig.EngineConfig,
		Subgraphs:          routerConfig.Subgraphs,
		RouterEngineConfig: routerEngineConfig,
		Reporter:           engineStats,
		Broker:             f.broker,
		Logger:             f.logger,
		Introspection:      true,
		InstanceData: types.InstanceData{
			HostName:      "localhost",
			ListenAddress: "4223",
		},
	}

	ecb := executor.ExecutorConfigurationBuilder{}

	ctx := context.Background()

	exec, pubsubProviders, err := ecb.Build(ctx, ecbParams)
	if err != nil {
		f.logger.Error("Failed to build executor configuration", zap.Error(err))
		return
	}

	f.pubsubProviders = pubsubProviders
	if pubSubStartupErr := f.startupProviders(ctx); pubSubStartupErr != nil {
		f.logger.Error("Failed to startup pubsub providers", zap.Error(pubSubStartupErr))
		return // Handle startup error
	}

	handler := fhandlers.NewFederationHandler(f.logger, exec)

	f.mu.Lock()
	f.handler = handler
	f.mu.Unlock()

	f.readyCh <- struct{}{}
}
