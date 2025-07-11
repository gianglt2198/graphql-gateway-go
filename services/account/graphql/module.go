package graphql

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/services/account/graphql/resolvers"
)

var Module = fx.Module("graphql",
	fx.Provide(resolvers.NewResolver),
)
