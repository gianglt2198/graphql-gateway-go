package utils

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/gianglt2198/federation-go/package/utils"
	"github.com/gianglt2198/federation-go/package/utils/system"
)

const (
	BadParameterErrorCode   = "BAD_PARAMETER_DATA"
	BadQueryErrorCode       = "BAD_QUERY_DATA"
	BadBodyErrorCode        = "BAD_BODY_DATA"
	InternalServerErrorCode = "INTERNAL_SERVER_ERROR"
	UnknownErrorCode        = "UNKNOWN"
	ValidationErrorCode     = "VALIDATION_ERROR"
	DuplicateDataErrorCode  = "DUPLICATE_DATA"
	NotFoundErrorCode       = "NOT_FOUND"
	InvalidUserErrorCode    = "INVALID_USER"
)

type ErrorMsg struct {
	Message    string   `json:"message"`
	Code       string   `json:"code"`
	Path       ast.Path `json:"-"`
	StackTrace []string `json:"stack"`
	HttpCode   int      `json:"-"`
}

type Error interface {
	Raise()
	RaiseWithHttpCode(httpCode int)
	RaiseWithContext(ctx context.Context)
	Error() string
	GetCode() string
	GetPath() ast.Path
	GetMessage() string
	GetStackTrace() []string
}

func NewError(message, code string) Error {
	return &ErrorMsg{
		Message: message,
		Code:    code,
	}
}

func (e *ErrorMsg) Raise() {
	e.StackTrace = system.GetStackTrace()
	panic(e)
}

func (e *ErrorMsg) RaiseWithContext(ctx context.Context) {
	e.StackTrace = system.GetStackTrace()
	e.Path = graphql.GetPath(ctx)
	panic(e)
}

func (e *ErrorMsg) RaiseWithHttpCode(httpCode int) {
	e.HttpCode = httpCode
	e.StackTrace = system.GetStackTrace()
	panic(e)
}

func (e *ErrorMsg) GetCode() string {
	return e.Code
}

func (e *ErrorMsg) GetPath() ast.Path {
	return e.Path
}

func (e *ErrorMsg) GetMessage() string {
	return e.Message
}

func (e *ErrorMsg) GetStackTrace() []string {
	return e.StackTrace
}

func (e *ErrorMsg) Error() string {
	return e.Message
}

func HandleGraphqlError(ctx context.Context, e error) *gqlerror.Error {
	err := graphql.DefaultErrorPresenter(ctx, e)
	if e != nil {
		var appErr Error
		if errors.As(e, &appErr) {
			err.Extensions = map[string]interface{}{
				"code":       appErr.GetCode(),
				"request_id": utils.GetRequestIDFromCtx(ctx),
			}
		} else {
			err.Extensions = map[string]interface{}{
				"code":       UnknownErrorCode,
				"request_id": utils.GetRequestIDFromCtx(ctx),
			}
		}
	}
	return err
}

func RecoverFunc(ctx context.Context, err any) error {
	if e, ok := err.(error); ok {
		return HandleGraphqlError(ctx, e)
	}
	return HandleGraphqlError(ctx, NewError("internal system error", InternalServerErrorCode))
}
