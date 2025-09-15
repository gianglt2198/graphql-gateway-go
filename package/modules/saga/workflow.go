package saga

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"go.uber.org/zap"

	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
	"github.com/gianglt2198/federation-go/package/utils/reflection"
)

const (
	WORKFLOW_INIT     = "init"
	WORKFLOW_SKIP     = "skip"
	WORKFLOW_SKIP_ALL = "skip_all"
)

type (
	State      map[string]any
	HandleFunc func(ctx context.Context, logger *logging.Logger, client graphql.Client) (any, error)
	StepFunc   func(state State) HandleFunc
	FinalFunc  func(ctx context.Context, state State) (*PlayResult, error)
	PlayResult struct {
		State State
		Error error
	}

	Step struct {
		Name         string
		Normal       StepFunc
		Compensation StepFunc
	}

	Saga[T any] struct {
		client    graphql.Client
		logger    *logging.Logger
		name      string
		steps     []Step
		cbSuccess FinalFunc
	}

	Workflow[R any] struct {
		Operation   any
		Function    any
		Variables   []any
		Skipable    bool
		SkipableAll bool
	}
)

func SagaBuilder[T any](logger *logging.Logger, client graphql.Client, name string) *Saga[T] {
	return &Saga[T]{
		client: client,
		logger: logger,
		name:   name,
		steps:  make([]Step, 0),
	}
}

func (b *Saga[T]) AddStep(step Step) *Saga[T] {
	b.steps = append(b.steps, step)
	return b
}

func (b *Saga[T]) AddSteps(steps ...Step) *Saga[T] {
	b.steps = append(b.steps, steps...)
	return b
}

func (b *Saga[T]) SetFinal(cb FinalFunc) *Saga[T] {
	b.cbSuccess = cb
	return b
}

func GetState[T any](state State, k string) *T {
	v, ok := state[k]
	if !ok || v == nil || v == WORKFLOW_SKIP || v == WORKFLOW_SKIP_ALL {
		return nil
	}

	t := v.(*T)

	return t
}

func (b *Saga[T]) Play(ctx context.Context, dto T) (*PlayResult, error) {
	stepNum := len(b.steps)

	sagaName := fmt.Sprintf("[%s]Play", b.name)
	b.logger.GetWrappedLogger(ctx).Info(sagaName, zap.Int("step_num", stepNum))

	if stepNum == 0 {
		b.logger.GetWrappedLogger(ctx).Info(sagaName, zap.String("status", "nothing to play"))
		return new(PlayResult), nil
	}

	compensationFuncs := make(map[string]StepFunc)
	state := make(map[string]any)

	state[WORKFLOW_INIT] = dto
	for _, activity := range b.steps {
		b.logger.GetWrappedLogger(ctx).Info(sagaName + ":" + activity.Name + "-normal")

		result, err := activity.Normal(state)(ctx, b.logger, b.client)
		if err != nil {
			if len(compensationFuncs) > 0 {
				b.logger.GetWrappedLogger(ctx).Info(sagaName + ":" + activity.Name + "-compensation")
				for k, f := range compensationFuncs {
					b.logger.GetWrappedLogger(ctx).Info(sagaName + ":" + k + "-compensation")
					_, cerr := f(state)(ctx, b.logger, b.client)
					if cerr != nil {
						b.logger.GetWrappedLogger(ctx).Error(sagaName+":"+k+"-compensation", zap.Error(cerr))
					}
				}
			}

			return nil, err
		}

		if skipAll, ok := result.(string); ok && skipAll == WORKFLOW_SKIP_ALL {
			break
		}

		state[activity.Name] = result
		if activity.Compensation != nil {
			compensationFuncs[activity.Name] = activity.Compensation
		}
	}

	if b.cbSuccess != nil {
		b.logger.GetWrappedLogger(ctx).Info(sagaName + "-final")
		return b.cbSuccess(ctx, state)
	}

	return &PlayResult{
		State: state,
	}, nil
}

func WorkflowFunc[R any](
	f func(state State) Workflow[R],
) StepFunc {
	return func(state State) HandleFunc {
		return func(ctx context.Context, logger *logging.Logger, client graphql.Client) (any, error) {
			data := f(state)
			if data.Skipable {
				return WORKFLOW_SKIP, nil
			}

			if data.SkipableAll {
				return WORKFLOW_SKIP_ALL, nil
			}

			var (
				res *R
				err error
			)

			if data.Operation != nil {
				params := []any{ctx, client}
				params = append(params, data.Variables...)
				res, err = reflection.CallFunctionWithError[*R](data.Operation, params...)
			}
			if data.Function != nil {
				params := []any{ctx}
				params = append(params, data.Variables...)
				res, err = reflection.CallFunctionWithError[*R](data.Function, params...)
			}
			if err != nil {
				return nil, err
			}
			return res, nil
		}
	}
}
