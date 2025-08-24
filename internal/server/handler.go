package server

import (
	"fmt"
	"io"

	"github.com/phungducminh/httpfromtcp/internal/request"
	"github.com/phungducminh/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func NewHandlerError(statusCode response.StatusCode, message string) *HandlerError {
	return &HandlerError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf("status code: %d, message = %s", e.StatusCode, e.Message)
}
