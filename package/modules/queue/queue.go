package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	credis "github.com/gianglt2198/federation-go/package/infras/cache/redis"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

type Queue struct {
	appConfig config.AppConfig

	log       *logging.Logger
	client    *asynq.Client
	inspector *asynq.Inspector
}

type QueueParams struct {
	fx.In

	AppConfig   config.AppConfig
	QueueConfig config.QueueConfig

	Logger *logging.Logger
	Redis  *credis.Redis
}

func NewQueue(params QueueParams) *Queue {
	if !params.QueueConfig.Enabled {
		return nil
	}

	r := params.Redis.GetClient()
	return &Queue{
		appConfig: params.AppConfig,
		log:       params.Logger,

		client:    asynq.NewClientFromRedisClient(r),
		inspector: asynq.NewInspectorFromRedisClient(r),
	}
}

func (q *Queue) applyDefaultOpts(opts []asynq.Option) []asynq.Option {
	return append(opts, asynq.Queue(q.appConfig.Name), asynq.Retention(24*time.Hour))
}

func (q *Queue) Enqueue(task *Task) (*asynq.TaskInfo, error) {
	opts := []asynq.Option{}
	q.applyDefaultOpts(opts)
	opts = append(opts, task.GetOptions()...)
	payload, err := json.Marshal(task.data)
	if err != nil {
		return nil, err
	}

	info, err := q.client.Enqueue(asynq.NewTask(task.name, payload), opts...)
	q.log.GetLogger().Info("Task info", zap.String("Task enqueued", task.name), zap.Any("data", task.data), zap.Any("options", task.opts))
	if err != nil {
		return nil, err
	}
	return info, err
}

func (q *Queue) EnqueueContext(ctx context.Context, task *Task) (*asynq.TaskInfo, error) {
	opts := []asynq.Option{}
	q.applyDefaultOpts(opts)
	opts = append(opts, task.GetOptions()...)
	payload, err := json.Marshal(task.data)
	if err != nil {
		return nil, err
	}

	info, err := q.client.EnqueueContext(ctx, asynq.NewTask(task.name, payload), opts...)
	q.log.GetWrappedLogger(ctx).Info("Task info", zap.String("Task enqueued", task.name), zap.Any("data", task.data), zap.Any("options", task.opts))
	if err != nil {
		return nil, err
	}
	return info, err
}
