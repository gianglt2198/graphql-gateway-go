package resolvers

import (
	"go.uber.org/fx"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gianglt2198/graphql-gateway-go/catalog/ent"
	"github.com/gianglt2198/graphql-gateway-go/catalog/graph/generated"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	log    *monitoring.AppLogger
	client *ent.Client
}

type ResolverParams struct {
	fx.In

	Log    *monitoring.AppLogger
	Client *ent.Client
}

// NewSchema creates a graphql executable schema.
func NewSchema(params ResolverParams) graphql.ExecutableSchema {
	return generated.NewExecutableSchema(generated.Config{
		Resolvers: &Resolver{
			log:    params.Log,
			client: params.Client,
		},
	})
}
