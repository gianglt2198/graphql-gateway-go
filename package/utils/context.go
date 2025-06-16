package utils

import (
	"context"

	"github.com/gianglt2198/federation-go/package/common"
)

func NewRequestID() string {
	return NewID(32, "req")
}

func GetRequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(common.KEY_REQUEST_ID).(string)
	if !ok {
		return ""
	}
	return requestID
}

func GetUserIDFromCtx(ctx context.Context) string {
	userID, ok := ctx.Value(common.KEY_AUTH_USER_ID).(string)
	if !ok {
		return ""
	}
	return userID
}

func GetTraceIDFromCtx(ctx context.Context) string {
	traceID, ok := ctx.Value(common.KEY_TRACE_ID).(string)
	if !ok {
		return ""
	}
	return traceID
}

func GetSpanIDFromCtx(ctx context.Context) string {
	spanID, ok := ctx.Value(common.KEY_SPAN_ID).(string)
	if !ok {
		return ""
	}
	return spanID
}

func GetValueByKeyFromCtx[T any](ctx context.Context, k common.KeyType) *T {
	if v, ok := ctx.Value(common.KeyType(k)).(*T); ok {
		return v
	}

	return nil
}

func ApplyTraceIDWithContext(ctx context.Context) (context.Context, string) {
	if traceID := GetTraceIDFromCtx(ctx); traceID != "" {
		return ctx, traceID
	}

	traceID := NewID(32, "trace")
	return context.WithValue(ctx, common.KEY_TRACE_ID, traceID), traceID
}

func ApplySpanIDWithContext(ctx context.Context) (context.Context, string) {
	if spanID := GetSpanIDFromCtx(ctx); spanID != "" {
		return ctx, spanID
	}

	spanID := NewID(32, "span")
	return context.WithValue(ctx, common.KEY_SPAN_ID, spanID), spanID
}

func ApplyRequestIDWithContext(ctx context.Context) (context.Context, string) {
	if requestID := GetRequestIDFromCtx(ctx); requestID != "" {
		return ctx, requestID
	}

	requestID := NewRequestID()
	return context.WithValue(ctx, common.KEY_REQUEST_ID, requestID), requestID
}

func ApplyValueByKeyWithCtx[T any](ctx context.Context, k common.KeyType, v *T) context.Context {
	return context.WithValue(ctx, k, v)
}
