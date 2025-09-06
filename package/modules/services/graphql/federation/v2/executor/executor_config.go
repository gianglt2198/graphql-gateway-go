package executor

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"github.com/wundergraph/graphql-go-tools/execution/graphql"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/apollocompatibility"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/datasource/introspection_datasource"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/plan"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"

	nodev1 "github.com/wundergraph/cosmo/router/gen/proto/wg/cosmo/node/v1"
	pubsub_datasource "github.com/wundergraph/cosmo/router/pkg/pubsub/datasource"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/package/infras/pubsub"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/types"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/loader"
	"github.com/gianglt2198/federation-go/package/modules/services/graphql/federation/v2/resolver"
)

type ExecutorConfigurationBuilder struct{}

type ExecutorConfigurationBuildParams struct {
	EngineConfig       *nodev1.EngineConfiguration
	Subgraphs          []*nodev1.Subgraph
	RouterEngineConfig *loader.RouterEngineConfiguration
	Reporter           resolve.Reporter
	InstanceData       types.InstanceData
	Broker             pubsub.Broker
	Logger             *monitoring.Logger
	Introspection      bool
}

func (b *ExecutorConfigurationBuilder) Build(ctx context.Context, params ExecutorConfigurationBuildParams) (*Executor, []pubsub_datasource.Provider, error) {
	planConfig, providers, err := b.buildPlannerConfiguration(ctx, &params)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build planner configuration: %w", err)
	}

	options := resolve.ResolverOptions{
		MaxConcurrency:                     params.RouterEngineConfig.Execution.MaxConcurrentResolvers,
		Debug:                              params.RouterEngineConfig.Execution.Debug.EnableResolverDebugging,
		Reporter:                           params.Reporter,
		PropagateSubgraphErrors:            params.RouterEngineConfig.SubgraphErrorPropagation.Enabled,
		PropagateSubgraphStatusCodes:       params.RouterEngineConfig.SubgraphErrorPropagation.PropagateStatusCodes,
		RewriteSubgraphErrorPaths:          params.RouterEngineConfig.SubgraphErrorPropagation.RewritePaths,
		OmitSubgraphErrorLocations:         params.RouterEngineConfig.SubgraphErrorPropagation.OmitLocations,
		OmitSubgraphErrorExtensions:        params.RouterEngineConfig.SubgraphErrorPropagation.OmitExtensions,
		AllowedErrorExtensionFields:        params.RouterEngineConfig.SubgraphErrorPropagation.AllowedExtensionFields,
		AttachServiceNameToErrorExtensions: params.RouterEngineConfig.SubgraphErrorPropagation.AttachServiceName,
		DefaultErrorExtensionCode:          params.RouterEngineConfig.SubgraphErrorPropagation.DefaultExtensionCode,
		AllowedSubgraphErrorFields:         params.RouterEngineConfig.SubgraphErrorPropagation.AllowedFields,
		AllowAllErrorExtensionFields:       params.RouterEngineConfig.SubgraphErrorPropagation.AllowAllExtensionFields,
		MaxRecyclableParserSize:            params.RouterEngineConfig.Execution.ResolverMaxRecyclableParserSize,
		MaxSubscriptionFetchTimeout:        params.RouterEngineConfig.Execution.SubscriptionFetchTimeout,
	}

	// this is the resolver, it's stateful and manages all the client connections, etc...
	resolver := resolve.New(ctx, options)

	var (
		// clientSchemaDefinition is the GraphQL Schema that is exposed from our API
		// it should be used for the introspection and query normalization/validation.
		clientSchemaDefinition *ast.Document
		// routerSchemaDefinition the GraphQL Schema that we use for planning the queries
		routerSchemaDefinition ast.Document
		report                 operationreport.Report
	)

	routerSchemaDefinition, report = astparser.ParseGraphqlDocumentString(params.EngineConfig.GraphqlSchema)
	if report.HasErrors() {
		return nil, providers, fmt.Errorf("failed to parse graphql schema from engine config: %w", report)
	}

	err = asttransform.MergeDefinitionWithBaseSchema(&routerSchemaDefinition)
	if err != nil {
		return nil, providers, fmt.Errorf("failed to merge graphql schema with base schema: %w", err)
	}

	if clientSchemaStr := params.EngineConfig.GetGraphqlClientSchema(); clientSchemaStr != "" {
		// The client schema is a subset of the router schema that does not include @inaccessible fields.
		// The client schema only exists if the federated schema includes @inaccessible directives or @tag directives

		clientSchema, report := astparser.ParseGraphqlDocumentString(clientSchemaStr)
		if report.HasErrors() {
			return nil, providers, fmt.Errorf("failed to parse graphql client schema from engine config: %w", report)
		}
		err = asttransform.MergeDefinitionWithBaseSchema(&clientSchema)
		if err != nil {
			return nil, providers, fmt.Errorf("failed to merge graphql client schema with base schema: %w", err)
		}
		clientSchemaDefinition = &clientSchema
	} else {
		// In the event that a client schema is not generated, the router schema is used in place of the client schema (e.g., for operation validation)

		clientSchemaDefinition = &routerSchemaDefinition
	}

	if params.Introspection {
		// by default, the engine doesn't understand how to resolve the __schema and __type queries
		// we need to add a special datasource for that
		// it takes the definition as the input and generates introspection data
		// datasource is attached to Query.__schema, Query.__type, __Type.fields and __Type.enumValues fields
		introspectionFactory, err := introspection_datasource.NewIntrospectionConfigFactory(clientSchemaDefinition)
		if err != nil {
			return nil, providers, fmt.Errorf("failed to create introspection config factory: %w", err)
		}
		fieldConfigs := introspectionFactory.BuildFieldConfigurations()
		// we need to add these fields to the config
		// otherwise the engine wouldn't know how to resolve them
		planConfig.Fields = append(planConfig.Fields, fieldConfigs...)
		dataSources := introspectionFactory.BuildDataSourceConfigurations()
		// finally, we add our data source for introspection to the existing data sources
		planConfig.DataSources = append(planConfig.DataSources, dataSources...)
	}

	var renameTypeNames []resolve.RenameTypeName

	// when applying namespacing, it's possible that we need to rename types
	// for that, we have to map the rename types config to the engine's rename type names
	for _, configuration := range planConfig.Types {
		if configuration.RenameTo != "" {
			renameTypeNames = append(renameTypeNames, resolve.RenameTypeName{
				From: []byte(configuration.RenameTo),
				To:   []byte(configuration.TypeName),
			})
		}
	}

	executionPlanCache, _ := lru.New(1000)

	schemaSDL := params.EngineConfig.GraphqlSchema

	schema, err := graphql.NewSchemaFromString(schemaSDL)
	if err != nil {
		return nil, providers, fmt.Errorf("failed to create schema from string: %w", err)
	}

	return &Executor{
		PlanConfig:         *planConfig,
		ClientSchema:       clientSchemaDefinition,
		RouterSchema:       &routerSchemaDefinition,
		Resolver:           resolver,
		RenameTypeNames:    renameTypeNames,
		executionPlanCache: executionPlanCache,
		apolloCompatibilityFlags: apollocompatibility.Flags{
			ReplaceInvalidVarError: true,
		},
		Schema: schema,
	}, providers, nil
}

func (b *ExecutorConfigurationBuilder) buildPlannerConfiguration(ctx context.Context, params *ExecutorConfigurationBuildParams) (*plan.Configuration, []pubsub_datasource.Provider, error) {
	// Implementation of the planner configuration building logic
	factory := resolver.NewDefaultFactoryResolver(ctx, params.Logger, true, params.InstanceData, params.Broker)

	loader := loader.NewLoader(ctx, factory, params.Logger)

	planConfig, providers, err := loader.Load(params.EngineConfig, params.Subgraphs, params.RouterEngineConfig, params.Introspection)

	return planConfig, providers, err
}
