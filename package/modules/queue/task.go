package queue

import "github.com/hibiken/asynq"

type Task struct {
	name string
	data any
	opts []asynq.Option
}

func NewTask[T any](name string) func(*T, ...asynq.Option) *Task {
	return func(data *T, opts ...asynq.Option) *Task {
		return &Task{
			name: name,
			data: data,
			opts: opts,
		}
	}
}

func (t Task) GetName() string {
	return t.name
}

func (t Task) GetData() any {
	return t.data
}

func (t Task) GetOptions() []asynq.Option {
	return t.opts
}
