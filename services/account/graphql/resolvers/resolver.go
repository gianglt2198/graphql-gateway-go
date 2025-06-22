package resolvers

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/gianglt2198/federation-go/package/infras/monitoring"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/generated"
	"github.com/gianglt2198/federation-go/services/account/internal/services"
	"go.uber.org/fx"
)

type Resolver struct {
	log *monitoring.Logger

	userService services.UserService
	authService services.AuthService
}

type ResolverParams struct {
	fx.In

	Log *monitoring.Logger

	UserService services.UserService
	AuthService services.AuthService
}

func NewResolver(params ResolverParams) graphql.ExecutableSchema {
	return generated.NewExecutableSchema(generated.Config{
		Resolvers: &Resolver{
			log: params.Log,

			userService: params.UserService,
			authService: params.AuthService,
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
