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

type Writer struct {
	wr io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		wr: w,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.wr.Write(p)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) (int, error) {
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

func (w *Writer) WriteHeaders(h *headers.Headers) (int, error) {
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

func (w *Writer) WriteBody(body []byte) (int, error) {
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

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	n1, err := w.wr.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		return 0, err
	}

	n2, err := w.wr.Write(p)
	if err != nil {
		return n1, err
	}

	n3, err := w.wr.Write([]byte("\r\n"))
	if err != nil {
		return n1 + n2, err
	}

	return n1 + n2 + n3, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.wr.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return 0, err
	}

	return n, err
}

func (w *Writer) WriteInternalServerError(err error, h *headers.Headers) {
	body := err.Error()
	h.Delete("Transfer-Encoding")
	h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteStatusLine(InternalServerError)
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}
