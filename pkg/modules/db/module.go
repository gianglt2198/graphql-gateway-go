package db

import (
	"fmt"

	"go.uber.org/fx"
)

func Module[T any](serviceName string, newClientFn NewClientFn[T]) fx.Option {
	return fx.Module(fmt.Sprintf("%v_db", serviceName),
		fx.Supply(
			fx.Private,
			fx.Annotate(newClientFn, fx.As(new(NewClientFn[T]))),
		),
		fx.Provide(NewDBClient[T]),
	)
}
