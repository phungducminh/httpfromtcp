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

func (h *HandlerError) Error() string {
	return fmt.Sprintf("status code: %d, message = %s", h.StatusCode, h.Message)
}

func (h *HandlerError) WriteTo(w io.Writer) (int64, error) {
	sln, err := response.WriteStatusLine(w, h.StatusCode)
	if err != nil {
		return 0, err
	}
	hn, err := response.WriteHeaders(w, response.GetDefaultHeaders(len(h.Message)))
	if err != nil {
		return int64(sln), err
	}
	bn, err := response.WriteBody(w, []byte(h.Message))
	if err != nil {
		return int64(sln + hn), err
	}

	return int64(sln + hn + bn), nil
}
