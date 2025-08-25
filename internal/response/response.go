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

func WriteStatusLine(w io.Writer, statusCode StatusCode) (int, error) {
	var err error
	var n int
	switch statusCode {
	case OK:
		n, err = w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case BadRequest:
		n, err = w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case InternalServerError:
		n, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		n, err = w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)))
	}

	return n, err
}

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) (int, error) {
	var err error
	var n int
	h.ForEach(func(key string, value string) {
		wn, werr := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		n += wn
		// only write the 1st error
		if err == nil {
			err = werr
		}
	})

	return n, err
}

func WriteBody(w io.Writer, body []byte) (int, error) {
	crn, err := w.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	bn, err := w.Write(body)
	if err != nil {
		return crn, err
	}
	return crn + bn, nil
}
