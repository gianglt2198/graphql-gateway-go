package graph

import (
	"github.com/gianglt2198/graphql-gateway-go/catalog/graph/resolvers"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("graph",
		fx.Provide(resolvers.NewSchema),
	)
}
