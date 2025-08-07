package loader

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/datasource/graphql_datasource"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/plan"

	"github.com/wundergraph/cosmo/router/gen/proto/wg/cosmo/common"
	nodev1 "github.com/wundergraph/cosmo/router/gen/proto/wg/cosmo/node/v1"
	"github.com/wundergraph/cosmo/router/pkg/config"
	pubsub_datasource "github.com/wundergraph/cosmo/router/pkg/pubsub/datasource"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	fpubsub "github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/types"
)

type FactoryResolver interface {
	ResolveGraphqlFactory(subgraphName string) (plan.PlannerFactory[graphql_datasource.Configuration], error)
	InstanceData() types.InstanceData
}

type RouterEngineConfiguration struct {
	Execution                config.EngineExecutionConfiguration
	Headers                  *config.HeaderRules
	Events                   config.EventsConfiguration
	SubgraphErrorPropagation config.SubgraphErrorPropagationConfiguration
}

type Loader struct {
	ctx      context.Context
	resolver FactoryResolver
	// includeInfo controls whether additional information like type usage and field usage is included in the plan definition
	logger *monitoring.Logger
}

func NewLoader(ctx context.Context, resolver FactoryResolver, logger *monitoring.Logger) *Loader {
	return &Loader{
		ctx:      ctx,
		resolver: resolver,
		logger:   logger,
	}
}

func (l *Loader) Load(engineConfig *nodev1.EngineConfiguration, subgraphs []*nodev1.Subgraph, routerEngineConfig *RouterEngineConfiguration, pluginsEnabled bool) (*plan.Configuration, []pubsub_datasource.Provider, error) {
	var outConfig plan.Configuration
	// attach field usage information to the plan
	outConfig.DefaultFlushIntervalMillis = engineConfig.DefaultFlushInterval
	for _, configuration := range engineConfig.FieldConfigurations {
		var args []plan.ArgumentConfiguration
		for _, argumentConfiguration := range configuration.ArgumentsConfiguration {
			arg := plan.ArgumentConfiguration{
				Name: argumentConfiguration.Name,
			}
			switch argumentConfiguration.SourceType {
			case nodev1.ArgumentSource_FIELD_ARGUMENT:
				arg.SourceType = plan.FieldArgumentSource
			case nodev1.ArgumentSource_OBJECT_FIELD:
				arg.SourceType = plan.ObjectFieldSource
			}
			args = append(args, arg)
		}
		fieldConfig := plan.FieldConfiguration{
			TypeName:  configuration.TypeName,
			FieldName: configuration.FieldName,
			Arguments: args,
		}
		outConfig.Fields = append(outConfig.Fields, fieldConfig)
	}

	for _, configuration := range engineConfig.TypeConfigurations {
		outConfig.Types = append(outConfig.Types, plan.TypeConfiguration{
			TypeName: configuration.TypeName,
			RenameTo: configuration.RenameTo,
		})
	}

	var providers []pubsub_datasource.Provider
	var pubSubDS []fpubsub.DataSourceConfigurationWithMetadata

	for _, in := range engineConfig.DatasourceConfigurations {
		var out plan.DataSource

		switch in.Kind {
		case nodev1.DataSourceKind_GRAPHQL:
			header := http.Header{}
			for s, httpHeader := range in.CustomGraphql.Fetch.Header {
				for _, value := range httpHeader.Values {
					header.Add(s, config.LoadStringVariable(value))
				}
			}

			fetchUrl := config.LoadStringVariable(in.CustomGraphql.Fetch.GetUrl())

			subscriptionUrl := config.LoadStringVariable(in.CustomGraphql.Subscription.Url)
			if subscriptionUrl == "" {
				subscriptionUrl = fetchUrl
			}

			graphqlSchema, err := l.LoadInternedString(engineConfig, in.CustomGraphql.GetUpstreamSchema())
			if err != nil {
				return nil, nil, fmt.Errorf("could not load GraphQL schema for data source %s: %w", in.Id, err)
			}

			customScalarTypeFields := make([]graphql_datasource.SingleTypeField, len(in.CustomGraphql.CustomScalarTypeFields))
			for i, v := range in.CustomGraphql.CustomScalarTypeFields {
				customScalarTypeFields[i] = graphql_datasource.SingleTypeField{
					TypeName:  v.TypeName,
					FieldName: v.FieldName,
				}
			}

			if in.CustomGraphql.Subscription.Protocol != nil {
				if *in.CustomGraphql.Subscription.Protocol != common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_WS {
					return nil, providers, fmt.Errorf("unsupported subscription protocol %q for data source %s", *in.CustomGraphql.Subscription.Protocol, in.Id)
				}
			}

			wsSubprotocol := "auto"
			if in.CustomGraphql.Subscription.WebsocketSubprotocol != nil {
				switch *in.CustomGraphql.Subscription.WebsocketSubprotocol {
				case common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_WS:
					wsSubprotocol = "graphql-ws"
				case common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_TRANSPORT_WS:
					wsSubprotocol = "graphql-transport-ws"
				case common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_AUTO:
					wsSubprotocol = "auto"
				}
			}

			schemaConfiguration, err := graphql_datasource.NewSchemaConfiguration(
				graphqlSchema,
				&graphql_datasource.FederationConfiguration{
					Enabled:    in.CustomGraphql.Federation.Enabled,
					ServiceSDL: in.CustomGraphql.Federation.ServiceSdl,
				},
			)
			if err != nil {
				return nil, providers, fmt.Errorf("error creating schema configuration for data source %s: %w", in.Id, err)
			}

			customConfiguration, err := graphql_datasource.NewConfiguration(graphql_datasource.ConfigurationInput{
				Fetch: &graphql_datasource.FetchConfiguration{
					URL:    fetchUrl,
					Method: in.CustomGraphql.Fetch.Method.String(),
					Header: header,
				},
				Subscription: &graphql_datasource.SubscriptionConfiguration{
					URL:           subscriptionUrl,
					WsSubProtocol: wsSubprotocol,
				},
				SchemaConfiguration:    schemaConfiguration,
				CustomScalarTypeFields: customScalarTypeFields,
			})
			if err != nil {
				return nil, providers, fmt.Errorf("error creating custom configuration for data source %s: %w", in.Id, err)
			}

			dataSourceName := l.subgraphName(subgraphs, in.Id)

			factory, err := l.resolver.ResolveGraphqlFactory(dataSourceName)
			if err != nil {
				return nil, providers, err
			}

			out, err = plan.NewDataSourceConfigurationWithName(
				in.Id,
				dataSourceName,
				factory,
				l.dataSourceMetaData(in),
				customConfiguration,
			)
			if err != nil {
				return nil, providers, fmt.Errorf("error creating data source configuration for data source %s: %w", in.Id, err)
			}

		case nodev1.DataSourceKind_PUBSUB:
			pubSubDS = append(pubSubDS, fpubsub.DataSourceConfigurationWithMetadata{
				Configuration: in,
				Metadata:      l.dataSourceMetaData(in),
			})
		default:
			return nil, providers, fmt.Errorf("unknown data source type %q", in.Kind)
		}

		if out != nil {
			outConfig.DataSources = append(outConfig.DataSources, out)
		}
	}

	factoryProviders, factoryDataSources, err := fpubsub.BuildProvidersAndDataSources(
		l.ctx,
		routerEngineConfig.Events,
		l.logger,
		pubSubDS,
		l.resolver.InstanceData().HostName,
		l.resolver.InstanceData().ListenAddress,
	)
	if err != nil {
		return nil, providers, err
	}

	if len(factoryProviders) > 0 {
		providers = append(providers, factoryProviders...)
	}

	if len(factoryDataSources) > 0 {
		outConfig.DataSources = append(outConfig.DataSources, factoryDataSources...)
	}

	return &outConfig, providers, nil
}

func (l *Loader) subgraphName(subgraphs []*nodev1.Subgraph, dataSourceID string) string {
	i := slices.IndexFunc(subgraphs, func(s *nodev1.Subgraph) bool {
		return s.Id == dataSourceID
	})

	if i != -1 {
		return subgraphs[i].Name
	}

	return ""
}

func (l *Loader) dataSourceMetaData(in *nodev1.DataSourceConfiguration) *plan.DataSourceMetadata {
	var d plan.DirectiveConfigurations = make([]plan.DirectiveConfiguration, 0, len(in.Directives))

	out := &plan.DataSourceMetadata{
		RootNodes:  make([]plan.TypeField, 0, len(in.RootNodes)),
		ChildNodes: make([]plan.TypeField, 0, len(in.ChildNodes)),
		Directives: &d,
		FederationMetaData: plan.FederationMetaData{
			Keys:     make([]plan.FederationFieldConfiguration, 0, len(in.Keys)),
			Requires: make([]plan.FederationFieldConfiguration, 0, len(in.Requires)),
			Provides: make([]plan.FederationFieldConfiguration, 0, len(in.Provides)),
		},
	}

	for _, node := range in.RootNodes {
		out.RootNodes = append(out.RootNodes, plan.TypeField{
			TypeName:           node.TypeName,
			FieldNames:         node.FieldNames,
			ExternalFieldNames: node.ExternalFieldNames,
		})
	}
	for _, node := range in.ChildNodes {
		out.ChildNodes = append(out.ChildNodes, plan.TypeField{
			TypeName:           node.TypeName,
			FieldNames:         node.FieldNames,
			ExternalFieldNames: node.ExternalFieldNames,
		})
	}
	for _, directive := range in.Directives {
		*out.Directives = append(*out.Directives, plan.DirectiveConfiguration{
			DirectiveName: directive.DirectiveName,
			RenameTo:      directive.DirectiveName,
		})
	}

	for _, keyConfiguration := range in.Keys {
		out.FederationMetaData.Keys = append(out.FederationMetaData.Keys, plan.FederationFieldConfiguration{
			TypeName:              keyConfiguration.TypeName,
			FieldName:             keyConfiguration.FieldName,
			SelectionSet:          keyConfiguration.SelectionSet,
			DisableEntityResolver: keyConfiguration.DisableEntityResolver,
		})
	}
	for _, providesConfiguration := range in.Provides {
		out.FederationMetaData.Provides = append(out.FederationMetaData.Provides, plan.FederationFieldConfiguration{
			TypeName:     providesConfiguration.TypeName,
			FieldName:    providesConfiguration.FieldName,
			SelectionSet: providesConfiguration.SelectionSet,
		})
	}
	for _, requiresConfiguration := range in.Requires {
		out.FederationMetaData.Requires = append(out.FederationMetaData.Requires, plan.FederationFieldConfiguration{
			TypeName:     requiresConfiguration.TypeName,
			FieldName:    requiresConfiguration.FieldName,
			SelectionSet: requiresConfiguration.SelectionSet,
		})
	}
	for _, entityInterfacesConfiguration := range in.EntityInterfaces {
		out.FederationMetaData.EntityInterfaces = append(out.FederationMetaData.EntityInterfaces, plan.EntityInterfaceConfiguration{
			InterfaceTypeName: entityInterfacesConfiguration.InterfaceTypeName,
			ConcreteTypeNames: entityInterfacesConfiguration.ConcreteTypeNames,
		})
	}
	for _, interfaceObjectConfiguration := range in.InterfaceObjects {
		out.FederationMetaData.InterfaceObjects = append(out.FederationMetaData.InterfaceObjects, plan.EntityInterfaceConfiguration{
			InterfaceTypeName: interfaceObjectConfiguration.InterfaceTypeName,
			ConcreteTypeNames: interfaceObjectConfiguration.ConcreteTypeNames,
		})
	}

	return out
}

func (l *Loader) LoadInternedString(engineConfig *nodev1.EngineConfiguration, str *nodev1.InternedString) (string, error) {
	key := str.GetKey()
	s, ok := engineConfig.StringStorage[key]
	if !ok {
		return "", fmt.Errorf("no string found for key %q", key)
	}
	return s, nil
}
