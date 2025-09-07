package tcnats

import (
	"context"

	nats "github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/gianglt2198/federation-go/package/common"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/tracing"
	"github.com/gianglt2198/federation-go/package/utils"
)

func OperationMiddleware(ctx context.Context, opeartionName string, handler func(context.Context, *nats.Msg) error) func(context.Context, *nats.Msg) error {
	if opeartionName == "" {
		opeartionName = "nats.operation"
	}

	return func(ctx context.Context, msg *nats.Msg) error {
		ctx = extractTraceContext(ctx, msg)

		tracer := tracing.Tracer("nats")
		ctx, span := tracer.Start(ctx, opeartionName,
			trace.WithSpanKind(trace.SpanKindConsumer),
		)

		ctx, requestID := utils.ApplyRequestIDWithContext(ctx)

		span.SetAttributes(attribute.String(string(common.KEY_REQUEST_ID), requestID))

		span.SetAttributes(attribute.String("nats.subject", msg.Subject))
		span.SetAttributes(attribute.Int("nats.payload_size", len(msg.Data)))

		_ = injectTraceContext(ctx, msg)
		err := handler(ctx, msg)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		span.End()

		return err
	}
}
