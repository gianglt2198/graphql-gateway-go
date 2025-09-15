package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

type ProcessorHandler struct {
	Pattern string
	Handler asynq.HandlerFunc
}

type ProcessorRunParams struct {
	fx.In

	QueueConfig config.QueueConfig

	Log      *logging.Logger
	Mux      *asynq.ServeMux
	Handlers []*ProcessorHandler `group:"processor"`
}

func Unmarshal[T any](task *asynq.Task) (*T, error) {
	var data T
	if err := json.Unmarshal(task.Payload(), &data); err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %v", err)
	}
	return &data, nil
}

func Handler[T any](pattern string, processTask func(ctx context.Context, payload *T, task *asynq.Task) error) *ProcessorHandler {
	return &ProcessorHandler{
		Pattern: pattern,
		Handler: func(ctx context.Context, task *asynq.Task) error {
			payload, err := Unmarshal[T](task)
			if err != nil {
				return err
			}
			return processTask(ctx, payload, task)
		},
	}
}

func RunProcessor(params ProcessorRunParams) {
	if !params.QueueConfig.Enabled {
		return
	}

	patterns := lo.Map(params.Handlers, func(handler *ProcessorHandler, index int) string {
		return handler.Pattern
	})
	params.Log.Info("Running processors", zap.Any("patterns", patterns))
	for _, handler := range params.Handlers {
		params.Mux.HandleFunc(handler.Pattern, handler.Handler)
	}
}
