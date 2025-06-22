package repos

import "go.uber.org/fx"

var Module = fx.Module("repos",
	fx.Provide(NewUserRepository),
	fx.Provide(NewSessionRepository),
)
