package queue

import "go.uber.org/fx"

var Module = []fx.Option{
	fx.Module("queue",
		fx.Provide(NewServer),
		fx.Provide(NewQueue),
		fx.Provide(NewRouter),
		fx.Invoke(RunServer, RunProcessor),
	),
}
