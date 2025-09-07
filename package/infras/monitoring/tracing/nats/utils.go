package tcnats

import (
	"context"

	nats "github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/propagation"

	"github.com/gianglt2198/federation-go/package/common"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/tracing"
	"github.com/gianglt2198/federation-go/package/utils"
)

func injectTraceContext(ctx context.Context, msg *nats.Msg) error {
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}

	carrier := propagation.MapCarrier{}
	tracing.Propagation().Inject(ctx, carrier)

	for key, val := range carrier {
		msg.Header.Set(key, val)
	}

	msg.Header.Set(string(common.KEY_REQUEST_ID), utils.GetRequestIDFromCtx(ctx))
	msg.Header.Set(string(common.KEY_AUTH_USER_ID), utils.GetUserIDFromCtx(ctx))
	return nil
}

func extractTraceContext(ctx context.Context, msg *nats.Msg) context.Context {
	carrier := propagation.MapCarrier{}
	if msg.Header != nil {
		for key, val := range msg.Header {
			if len(val) > 0 {
				carrier[key] = val[0]
			}
		}
	}

	ctx = tracing.Propagation().Extract(ctx, carrier)
	ctx = context.WithValue(ctx, common.KEY_REQUEST_ID, msg.Header.Get(string(common.KEY_REQUEST_ID)))
	ctx = context.WithValue(ctx, common.KEY_AUTH_USER_ID, msg.Header.Get(string(common.KEY_AUTH_USER_ID)))
	return ctx
}
