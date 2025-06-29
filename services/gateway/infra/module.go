package infra

import (
	"github.com/gianglt2198/federation-go/services/gateway/internal/app"
	"go.uber.org/fx"
)

var Module = fx.Module("infra",
	fx.Provide(app.New),
)
