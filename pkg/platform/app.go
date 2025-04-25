package platform

import (
	"github.com/gianglt2198/graphql-gateway-go/pkg"
	"go.uber.org/fx"
)

type application struct {
	options []fx.Option
}

type Application interface {
	Start(hooks ...fx.Hook)
}

func buildOptions(config *pkg.Config) {

}
