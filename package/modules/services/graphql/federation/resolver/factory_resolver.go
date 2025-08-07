package resolver

import (
	"context"
	"net/http"
	"time"

	"github.com/jensneuse/abstractlogger"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/datasource/graphql_datasource"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/plan"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/types"
	"github.com/gianglt2198/federation-go/package/modules/services/http/transports"
)

type DefaultFactoryResolver struct {
	logger *monitoring.Logger

	engineCtx context.Context

	httpClient         *http.Client
	streamingClient    *http.Client
	subscriptionClient graphql_datasource.GraphQLSubscriptionClient

	factoryLogger abstractlogger.Logger
	instanceData  types.InstanceData
}

func NewDefaultFactoryResolver(
	ctx context.Context,
	logger *monitoring.Logger,
	enableNetPoll bool,
	instanceData types.InstanceData,
	broker pubsub.Broker,
) *DefaultFactoryResolver {
	// Create HTTP client with custom transport for NATS support
	transport := transports.NewNatsTransport(transports.NatsTransportParams{
		Upstream: http.DefaultTransport.(*http.Transport),
		Logger:   logger,
		Broker:   broker,
	})

	defaultHTTPClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	streamingClient := &http.Client{
		Transport: transport,
	}

	var factoryLogger abstractlogger.Logger
	if logger != nil {
		factoryLogger = abstractlogger.NewZapLogger(logger.GetLogger(), abstractlogger.DebugLevel)
	}

	var netPollConfig graphql_datasource.NetPollConfiguration

	netPollConfig.ApplyDefaults()

	netPollConfig.Enable = enableNetPoll

	options := []graphql_datasource.Options{
		graphql_datasource.WithLogger(factoryLogger),
		graphql_datasource.WithNetPollConfiguration(netPollConfig),
	}

	subscriptionClient := graphql_datasource.NewGraphQLSubscriptionClient(
		defaultHTTPClient,
		streamingClient,
		ctx,
		options...,
	)

	return &DefaultFactoryResolver{
		logger:        logger,
		factoryLogger: factoryLogger,
		engineCtx:     ctx,

		streamingClient:    streamingClient,
		subscriptionClient: subscriptionClient,

		httpClient: defaultHTTPClient,

		instanceData: instanceData,
	}
}

func (d *DefaultFactoryResolver) ResolveGraphqlFactory(subgraphName string) (plan.PlannerFactory[graphql_datasource.Configuration], error) {
	return graphql_datasource.NewFactory(d.engineCtx, d.httpClient, d.subscriptionClient)
}

func (d *DefaultFactoryResolver) InstanceData() types.InstanceData {
	return d.instanceData
}
