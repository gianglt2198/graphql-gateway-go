package internal

import (
	"github.com/gianglt2198/federation-go/services/gateway/internal/app"
	"go.uber.org/fx"
)

// Module provides all gateway components for dependency injection
var Module = fx.Module("gateway",
	fx.Provide(app.New),
)
