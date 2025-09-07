package resolvers

import (
	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"

	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/generated"
	"github.com/gianglt2198/federation-go/services/catalog/internal/services"
)

type Resolver struct {
	log *logging.Logger

	categoryService services.CategoryService
	productService  services.ProductService
}

type ResolverParams struct {
	fx.In

	Log *logging.Logger

	CategoryService services.CategoryService
	ProductService  services.ProductService
}

func NewResolver(params ResolverParams) graphql.ExecutableSchema {
	return generated.NewExecutableSchema(generated.Config{
		Resolvers: &Resolver{
			log: params.Log,

			categoryService: params.CategoryService,
			productService:  params.ProductService,
		},
	})
}

type (
	queryResolver    struct{ *Resolver }
	mutationResolver struct{ *Resolver }
	entityResolver   struct{ *Resolver }
)

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }
