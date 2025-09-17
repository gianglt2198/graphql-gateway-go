package scheduler

import (
	"context"
	"encoding/json"

	"github.com/gianglt2198/federation-go/package/config"
	credis "github.com/gianglt2198/federation-go/package/infras/cache/redis"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/modules/queue"
	"github.com/hibiken/asynq"
	"go.uber.org/fx"
)

type Schedulable func(*Scheduler) error

type Scheduler struct {
	client *asynq.Scheduler
}

type SchedulerParams struct {
	fx.In

	AppConfig       config.AppConfig
	SchedulerConfig config.SchedulerConfig

	Logger *logging.Logger
	Redis  *credis.Redis
}

func NewScheduler(params SchedulerParams) *Scheduler {
	if !params.SchedulerConfig.Enabled {
		return nil
	}

	return &Scheduler{
		client: asynq.NewSchedulerFromRedisClient(
			params.Redis.GetClient(),
			&asynq.SchedulerOpts{
				Logger: params.Logger.Asynq(),
			},
		),
	}
}

func Schedule(cronspec string, task *queue.Task) Schedulable {
	return func(s *Scheduler) error {
		payload, err := json.Marshal(task.GetData())
		if err != nil {
			return err
		}
		_, err = s.client.Register(cronspec, asynq.NewTask(task.GetName(), payload), task.GetOptions()...)
		return err
	}
}

func Run(log *logging.Logger, lifecycle fx.Lifecycle, scheduler *Scheduler) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if scheduler == nil {
				return nil
			}

			log.Info("[OnStart] Scheduler running.....")
			errChan := make(chan error)
			go func() {
				defer close(errChan)
				if err := scheduler.client.Run(); err != nil {
					errChan <- err
				}
			}()
			select {
			case err := <-errChan:
				return err
			default:
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			scheduler.client.Shutdown()
			return nil
		},
	})
}
