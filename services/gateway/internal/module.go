package internal

import (
	"go.uber.org/fx"

	"github.com/gianglt2198/federation-go/services/gateway/internal/app"
)

// Module provides all gateway components for dependency injection
var Module = fx.Module("gateway",
	fx.Provide(app.New),
	fx.Invoke(app.Run),
)
