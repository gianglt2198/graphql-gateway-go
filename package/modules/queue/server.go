package queue

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	credis "github.com/gianglt2198/federation-go/package/infras/cache/redis"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

type QueueServer struct {
	servers []*asynq.Server
}

type QueueServerParams struct {
	fx.In

	AppConfig   config.AppConfig
	QueueConfig config.QueueConfig

	Logger *logging.Logger
	Redis  *credis.Redis
}

func NewServer(params QueueServerParams) *QueueServer {
	if !params.QueueConfig.Enabled {
		return nil
	}

	if len(params.QueueConfig.Providers) == 0 {
		params.QueueConfig.Providers = []config.ProviderConfig{
			{
				Priorities: []config.PriorityConfig{
					{
						Name:     params.AppConfig.Name,
						Priority: 10,
					},
				},
				Concurrency: 10,
			},
		}
	}

	res := &QueueServer{
		servers: make([]*asynq.Server, 0, len(params.QueueConfig.Providers)),
	}

	for _, provider := range params.QueueConfig.Providers {
		params.Logger.GetLogger().Info("Queue server initilaizing ....", zap.Any("priorities", provider.Priorities), zap.Int("concurrency", provider.Concurrency))

		res.servers = append(res.servers, asynq.NewServerFromRedisClient(
			params.Redis.GetClient(),
			asynq.Config{
				Logger:      params.Logger.Asynq(),
				Concurrency: provider.Concurrency,
				Queues: lo.SliceToMap(provider.Priorities, func(q config.PriorityConfig) (string, int) {
					return q.Name, q.Priority
				}),
				ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
					params.Logger.Error("error processing task", zap.Error(err), zap.Any("task", task))
				}),
			},
		))
	}

	return res
}

func NewRouter() *asynq.ServeMux {
	return asynq.NewServeMux()
}

func RunServer(log *logging.Logger, lifecycle fx.Lifecycle, servers *QueueServer, mux *asynq.ServeMux) {
	if servers == nil {
		return
	}
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errChan := make(chan error)
			log.Info("[OnStart] Queue running with servers", zap.Int("servers", len(servers.servers)))
			for _, srv := range servers.servers {
				go func() {
					defer close(errChan)
					if err := srv.Run(mux); err != nil {
						errChan <- err
					}
				}()
			}
			select {
			case err := <-errChan:
				log.Error(err.Error())
				return err
			default:
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			for _, srv := range servers.servers {
				srv.Shutdown()
			}
			return nil
		},
	})
}
