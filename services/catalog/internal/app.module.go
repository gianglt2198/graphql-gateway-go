package internal

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("app") // Graphql
}
