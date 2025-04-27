package service

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gianglt2198/graphql-gateway-go/pkg/utils/system"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	BadParameterErrorCode   = "BAD_PARAMETER_DATA"
	BadQueryErrorCode       = "BAD_QUERY_DATA"
	BadBodyErrorCode        = "BAD_BODY_DATA"
	InternalServerErrorCode = "INTERNAL_SERVER_ERROR"
	UnknownErrorCode        = "UNKNOWN"
	ValidationErrorCode     = "VALIDATION_ERROR"
)

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

type ErrorResponse struct {
	Err Error `json:"error"`
}

type ErrorMsg struct {
	Message    string   `json:"message"`
	Code       string   `json:"code"`
	Path       ast.Path `json:"-"`
	StackTrace []string `json:"stack"`
	HttpCode   int      `json:"-"`
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
