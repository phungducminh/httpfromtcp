package server

import (
	"fmt"

	"github.com/phungducminh/httpfromtcp/internal/request"
	"github.com/phungducminh/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

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

func (h *HandlerError) WriteTo(w *response.Writer) (int64, error) {
	sln, err := w.WriteStatusLine(h.StatusCode)
	if err != nil {
		return 0, err
	}
	hn, err := w.WriteHeaders(response.GetDefaultHeaders(len(h.Message)))
	if err != nil {
		return int64(sln), err
	}
	bn, err := w.WriteBody([]byte(h.Message))
	if err != nil {
		return int64(sln + hn), err
	}

	return int64(sln + hn + bn), nil
}
