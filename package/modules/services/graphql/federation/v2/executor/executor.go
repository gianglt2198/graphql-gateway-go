package executor

import (
	"context"
	"errors"
	"io"

	lru "github.com/hashicorp/golang-lru"

	"github.com/wundergraph/graphql-go-tools/execution/graphql"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/apollocompatibility"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/plan"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/engine/resolve"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/pool"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/variablesvalidation"
)

type Executor struct {
	PlanConfig plan.Configuration
	// ClientSchema is the GraphQL Schema that is exposed from our API
	// it is used for the introspection and query normalization/validation.
	ClientSchema *ast.Document
	// RouterSchema the GraphQL Schema that we use for planning the queries
	RouterSchema             *ast.Document
	Schema                   *graphql.Schema
	Resolver                 *resolve.Resolver
	RenameTypeNames          []resolve.RenameTypeName
	executionPlanCache       *lru.Cache
	apolloCompatibilityFlags apollocompatibility.Flags
}

func (e *Executor) Execute(ctx context.Context, operation *graphql.Request, writer resolve.SubscriptionResponseWriter) error {
	if err := e.normalizeOperation(operation); err != nil {
		return err
	}

	execContext := newInternalExecutionContext()
	execContext.prepare(ctx, operation.Variables, operation.InternalRequest())

	var report operationreport.Report
	cachedPlan := e.getCachedPlan(execContext, operation.Document(), e.RouterSchema, operation.OperationName, &report)
	if report.HasErrors() {
		return report
	}

	switch p := cachedPlan.(type) {
	case *plan.SynchronousResponsePlan:
		_, err := e.Resolver.ResolveGraphQLResponse(execContext.resolveContext, p.Response, nil, writer)
		return err
	case *plan.SubscriptionResponsePlan:
		return e.Resolver.ResolveGraphQLSubscription(execContext.resolveContext, p.Response, writer)
	default:
		return errors.New("execution of operation is not possible")
	}
}

func (e *Executor) normalizeOperation(operation *graphql.Request) error {
	normalize := !operation.IsNormalized()
	if normalize {
		// Normalize the operation, but extract variables later so ValidateForSchema can return correct error messages for bad arguments.
		result, err := operation.Normalize(e.Schema,
			astnormalization.WithRemoveFragmentDefinitions(),
			astnormalization.WithRemoveUnusedVariables(),
			astnormalization.WithInlineFragmentSpreads(),
		)
		if err != nil {
			return err
		} else if !result.Successful {
			return result.Errors
		}
		normalize = true
	}

	// Validate the operation against the schema.
	if result, err := operation.ValidateForSchema(e.Schema); err != nil {
		return err
	} else if !result.Valid {
		return result.Errors
	}

	if normalize {
		// Normalize the operation again, this time just extracting additional variables from arguments.
		result, err := operation.Normalize(e.Schema,
			astnormalization.WithExtractVariables(),
		)
		if err != nil {
			return err
		} else if !result.Successful {
			return result.Errors
		}
	}

	// Validate user-supplied and extracted variables against the operation.
	if len(operation.Variables) > 0 && operation.Variables[0] == '{' {
		validator := variablesvalidation.NewVariablesValidator(variablesvalidation.VariablesValidatorOptions{
			ApolloCompatibilityFlags: e.apolloCompatibilityFlags,
		})
		if err := validator.Validate(operation.Document(), e.RouterSchema, operation.Variables); err != nil {
			return err
		}
	}
	return nil
}

func (e *Executor) getCachedPlan(ctx *internalExecutionContext, operation, definition *ast.Document, operationName string, report *operationreport.Report) plan.Plan {
	hash := pool.Hash64.Get()
	hash.Reset()
	defer pool.Hash64.Put(hash)
	err := astprinter.Print(operation, hash)
	if err != nil {
		report.AddInternalError(err)
		return nil
	}

	cacheKey := hash.Sum64()

	if cached, ok := e.executionPlanCache.Get(cacheKey); ok {
		if p, ok := cached.(plan.Plan); ok {
			return p
		}
	}

	planner, _ := plan.NewPlanner(e.PlanConfig)
	planResult := planner.Plan(operation, definition, operationName, report)
	if report.HasErrors() {
		return nil
	}

	p := ctx.postProcessor.Process(planResult)
	e.executionPlanCache.Add(cacheKey, p)
	return p
}

func (e *Executor) ExecuteSubscription(ctx context.Context, operation *graphql.Request, writer resolve.SubscriptionResponseWriter, id resolve.SubscriptionIdentifier) (plan.Plan, error) {
	if err := e.normalizeOperation(operation); err != nil {
		return nil, err
	}

	execContext := newInternalExecutionContext()
	execContext.prepare(ctx, operation.Variables, operation.InternalRequest())

	var report operationreport.Report
	cachedPlan := e.getCachedPlan(execContext, operation.Document(), e.RouterSchema, operation.OperationName, &report)
	if report.HasErrors() {
		return nil, report
	}

	switch p := cachedPlan.(type) {
	case *plan.SynchronousResponsePlan:
		_, err := e.Resolver.ResolveGraphQLResponse(execContext.resolveContext, p.Response, nil, writer)
		if err != nil {
			return p, err
		}
	case *plan.SubscriptionResponsePlan:
		e.Resolver.SetAsyncErrorWriter(e)
		return p, e.Resolver.AsyncResolveGraphQLSubscription(execContext.resolveContext, p.Response, writer, id)
	}

	return cachedPlan, nil
}

func (e *Executor) WriteError(ctx *resolve.Context, err error, res *resolve.GraphQLResponse, w io.Writer) {
	_, _ = w.Write([]byte(err.Error()))
}
