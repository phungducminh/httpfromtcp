package response

import (
	"fmt"
	"io"

	"github.com/phungducminh/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error
	switch statusCode {
	case OK:
		_, err = w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case BadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case InternalServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		_, err = w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)))
	}

	return err
}

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	var err error
	h.ForEach(func(key string, value string) {
		_, werr := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		// only write the 1st error
		if err == nil {
			err = werr
		}
	})

	return err
}

func WriteBody(w io.Writer, body string) error {
	_, err := w.Write([]byte("\r\n" + body))
	return err
}
