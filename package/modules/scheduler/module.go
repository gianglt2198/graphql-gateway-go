package scheduler

import "go.uber.org/fx"

var Module = []fx.Option{
	fx.Module("scheduler",
		fx.Provide(NewScheduler),
		fx.Invoke(Run),
	),
}
