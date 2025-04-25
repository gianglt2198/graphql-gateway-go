package utils

import (
	"context"

	"github.com/gianglt2198/graphql-gateway-go/pkg/common"
)

func GetRequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(common.KEY_REQUEST_ID).(string)
	if !ok {
		return ""
	}
	return requestID
}

func NewRequestID() string {
	return NewID(32, "req")
}

func GetUserIDFromCtx(ctx context.Context) string {
	userID := ctx.Value(common.KEY_AUTH_USER_ID)

	if userID == nil {
		return ""
	}

	return userID.(string)
}
