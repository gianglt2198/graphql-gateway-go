package graphql

import (
	"github.com/gianglt2198/federation-go/services/account/graphql/resolvers"
	"go.uber.org/fx"
)

var Module = fx.Module("graphql",
	fx.Provide(resolvers.NewResolver),
)
