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

func newRequest() *Request {
	return &Request{
		state:   Initialized,
		Headers: headers.NewHeaders(),
		Body:    nil,
	}
}

// TODO: @minh move the logic down
// parse return number of bytes read and error
// the number of bytes read will be used for moving buffer
func (r *Request) parse(p []byte, eof bool) (int, error) {
	readN := 0
	for {
		switch r.state {
		case Done:
			return readN, nil
		case Error:
			return readN, nil
		case Initialized:
			r.state = ParsingRequestLine
		case ParsingRequestLine:
			rl, n, err := parseRequestLine(p[readN:])
			if err != nil {
				r.state = Error
				return readN, err
			}
			if rl == nil {
				return readN, nil
			}

			r.RequestLine = *rl
			r.state = ParsingHeaders
			readN += n
		case ParsingHeaders:
			s := p[readN:]
			n, done, err := r.Headers.Parse(s)
			if err != nil {
				r.state = Error
				return readN, err
			}
			if !done {
				return readN, nil
			}

			r.state = ParsingBody
			readN += n
		case ParsingBody:
			body, n, err := r.parseRequestBody(p[readN:], eof)
			if err != nil {
				r.state = Error
				return readN, err
			}
			if body == nil {
				return readN, nil
			}

			r.Body = body
			r.state = Done
			readN += n
		default:
			panic(fmt.Sprintf("unexpected request.RequestState: %#v", r.state))
		}
	}
}

func (r *Request) done() bool {
	return r.state == Done || r.state == Error
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	// TODO: overflow 1024 bytes
	buf := make([]byte, 1024)
	end := 0

	// buf[:end] denote the available buffer to be parsed by request
	for !req.done() {
		eof := false
		// read a chunk
		n, err := reader.Read(buf[end:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = true
			} else {
				return nil, err
			}
		}

		end += n
		readN, err := req.parse(buf[:end], eof)
		if err != nil {
			return nil, err
		}

		// shift available buffer to left
		copy(buf, buf[readN:end])
		end -= readN
	}

	return req, nil
}

// HTTP-version  = HTTP-name "/" DIGIT "." DIGIT
// HTTP-name     = %s"HTTP"
// request-line  = method SP request-target SP HTTP-version
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	lsi := bytes.Index(data, ls)
	if lsi == -1 {
		// not enough data for parsing
		return nil, 0, nil
	}
	data = data[:lsi]
	parts := bytes.Split(data, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrMalformedRequestLine
	}

	method := parts[0]
	if !slices.Equal(method, bytes.ToUpper(method)) {
		return nil, 0, ErrMalformedRequestLine
	}

	requestTarget := parts[1]
	if len(requestTarget) == 0 || bytes.Index(requestTarget, []byte("/")) != 0 {
		return nil, 0, ErrMalformedRequestLine
	}

	httpVersion := parts[2]
	bsi := bytes.Index(httpVersion, []byte("/"))
	version := string(httpVersion[bsi+1:])
	if string(httpVersion[:bsi]) != "HTTP" || version != "1.1" {
		return nil, 0, ErrMalformedRequestLine
	}

	rl := &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(requestTarget),
		Method:        string(method),
	}
	return rl, len(data) + len(ls), nil
}

func (r *Request) parseRequestBody(data []byte, eof bool) ([]byte, int, error) {
	cl := r.Headers.Get("content-length")
	if cl == "" {
		return []byte{}, 0, nil
	}

	length, err := strconv.ParseInt(cl, 10, 32)
	if err != nil || length < 0 {
		return nil, 0, ErrMalformedRequestHeaders
	}

	if eof && len(data) != int(length) {
		return nil, 0, ErrMalformedRequestBody
	}

	if len(data) < int(length) {
		// not enough data for request body
		return nil, 0, nil
	}

	if len(data) > int(length) {
		// mismatch content-length value and body length
		return nil, 0, ErrMalformedRequestBody
	}

	return data, len(data), nil
}
