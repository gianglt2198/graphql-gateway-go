package resolvers

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/services/catalog/generated/graph/generated"
	"github.com/gianglt2198/federation-go/services/catalog/internal/services"
	"go.uber.org/fx"
)

type Resolver struct {
	log *monitoring.Logger

	categoryService services.CategoryService
	productService  services.ProductService
}

type ResolverParams struct {
	fx.In

	Log *monitoring.Logger

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
)

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }
