package psnats

import (
	"context"

	nats "github.com/nats-io/nats.go"
)

type NatsMiddleware func(context.Context, string, func(context.Context, *nats.Msg) error) func(context.Context, *nats.Msg) error

func (n *natsProvider) Middleware(ctx context.Context, operation string, handler func(context.Context, *nats.Msg) error, ms ...NatsMiddleware) func(context.Context, *nats.Msg) error {
	h := handler
	// middleware are applied in reverse; this makes the first middleware
	// in the slice the outermost i.e. first to enter, last to exit
	// given: store, A, B, C
	// result: A(B(C(store)))
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](ctx, operation, h)
	}

	return h
}
