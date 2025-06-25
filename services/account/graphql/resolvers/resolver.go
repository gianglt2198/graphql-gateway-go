package resolvers

import (
	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/package/infras/monitoring"

	"github.com/gianglt2198/federation-go/services/account/generated/ent"
	"github.com/gianglt2198/federation-go/services/account/generated/graph/generated"
	"github.com/gianglt2198/federation-go/services/account/internal/services"
)

type Resolver struct {
	log *monitoring.Logger
	db  *ent.Client

	userService services.UserService
	authService services.AuthService
}

type ResolverParams struct {
	fx.In

	Log *monitoring.Logger
	Db  *ent.Client

	UserService services.UserService
	AuthService services.AuthService
}

func NewResolver(params ResolverParams) graphql.ExecutableSchema {
	return generated.NewExecutableSchema(generated.Config{
		Resolvers: &Resolver{
			log: params.Log,
			db:  params.Db,

			userService: params.UserService,
			authService: params.AuthService,
		},
	})
}

type (
	entityResolver   struct{ *Resolver }
	queryResolver    struct{ *Resolver }
	mutationResolver struct{ *Resolver }
)

func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }
