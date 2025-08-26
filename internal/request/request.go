package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"

	"github.com/phungducminh/httpfromtcp/internal/headers"
)

var ls = []byte("\r\n")

var (
	ErrMalformedRequest                     = fmt.Errorf("request: malformed request")
	ErrMalformedRequestLine                 = fmt.Errorf("request: malformed request line")
	ErrMalformedRequestHeaders              = fmt.Errorf("request: malformed request headers")
	ErrMalformedRequestHeadersContentLength = fmt.Errorf("request: malformed request headers content-length")
	ErrMalformedRequestBody                 = fmt.Errorf("request: malformed request body")
)

type RequestState string

const (
	Initialized        RequestState = "Initialized"
	Error              RequestState = "Error"
	Done               RequestState = "Done"
	ParsingRequestLine RequestState = "ParsingRequestLine"
	ParsingHeaders     RequestState = "ParsingHeaders"
	ParsingBody        RequestState = "ParsingBody"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
	Body          []byte
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       RequestState
}

func RequestFromReader(r io.Reader) (*Request, error) {
	req := newRequest()
	// TODO: overflow 1024 bytes
	b := make([]byte, 1024)
	end := 0

	// buf[:end] denote the available buffer to be parsed by request
	for !req.done() {
		eof := false
		// read a chunk
		rn, err := r.Read(b[end:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = true
			} else {
				return nil, err
			}
		}

		end += rn
		pn, err := req.parse(b[:end], eof)
		if err != nil {
			return nil, err
		}

		// shift available buffer to left
		copy(b, b[pn:end])
		end -= pn

		if eof && !req.done() {
			// EOF but parsing is not yet completed, there must be parsing 
			// implementation error
			return nil, fmt.Errorf("request: expect parsing completed after received EOF")
		}
	}

	return req, nil
}

func newRequest() *Request {
	return &Request{
		state: Initialized,
	}
}

// parse return number of bytes read and error
// the number of bytes read will be used for moving buffer
func (r *Request) parse(p []byte, eof bool) (int, error) {
	rn := 0
	for {
		switch r.state {
		case Done:
			return rn, nil
		case Error:
			return rn, nil
		case Initialized:
			r.state = ParsingRequestLine
		case ParsingRequestLine:
			rl, n, err := parseRequestLine(p[rn:])
			if err != nil {
				r.state = Error
				return rn, err
			}
			if rl == nil {
				return rn, nil
			}

			r.RequestLine = *rl
			r.state = ParsingHeaders
			rn += n
		case ParsingHeaders:
			h, n, err := headers.Parse(p[rn:], eof)
			if err != nil {
				r.state = Error
				return rn, err
			}
			if h == nil {
				return rn, nil
			}

			r.Headers = h
			r.state = ParsingBody
			rn += n
		case ParsingBody:
			body, n, err := parseRequestBody(p[rn:], eof, r.Headers)
			if err != nil {
				r.state = Error
				return rn, err
			}
			if body == nil {
				return rn, nil
			}

			r.Body = body
			r.state = Done
			rn += n
		default:
			panic(fmt.Sprintf("unexpected request.RequestState: %#v", r.state))
		}
	}
}

func (r *Request) done() bool {
	return r.state == Done || r.state == Error
}

func parseRequestBody(p []byte, eof bool, h *headers.Headers) ([]byte, int, error) {
	cl := h.Get("content-length")
	if cl == "" {
		return []byte{}, 0, nil
	}

	n, err := strconv.ParseInt(cl, 10, 32)
	if err != nil || n < 0 {
		return nil, 0, ErrMalformedRequestHeaders
	}

	if eof && len(p) != int(n) {
		return nil, 0, ErrMalformedRequestBody
	}

	if len(p) < int(n) {
		// not enough data for request body
		return nil, 0, nil
	}

	if len(p) > int(n) {
		// mismatch content-length value and body length
		return nil, 0, ErrMalformedRequestBody
	}

	return p, len(p), nil
}

// HTTP-version  = HTTP-name "/" DIGIT "." DIGIT
// HTTP-name     = %s"HTTP"
// request-line  = method SP request-target SP HTTP-version
func parseRequestLine(p []byte) (*RequestLine, int, error) {
	i := bytes.Index(p, ls)
	if i == -1 {
		// not enough data for parsing
		return nil, 0, nil
	}
	p = p[:i]
	parts := bytes.Split(p, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrMalformedRequestLine
	}

	method := parts[0]
	if !slices.Equal(method, bytes.ToUpper(method)) {
		return nil, 0, ErrMalformedRequestLine
	}

	target := parts[1]
	if len(target) == 0 || bytes.Index(target, []byte("/")) != 0 {
		return nil, 0, ErrMalformedRequestLine
	}

	httpVersion := parts[2]
	i = bytes.Index(httpVersion, []byte("/"))
	version := string(httpVersion[i+1:])
	if string(httpVersion[:i]) != "HTTP" || version != "1.1" {
		return nil, 0, ErrMalformedRequestLine
	}

	rl := &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(target),
		Method:        string(method),
	}
	return rl, len(p) + len(ls), nil
}
