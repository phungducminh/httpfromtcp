package request

import (
	"bytes"
	"fmt"
	"io"
	"slices"
)

var ls = []byte("\r\n")

var (
	ErrMalformedRequest     = fmt.Errorf("request: malformed request")
	ErrMalformedRequestLine = fmt.Errorf("request: malformed request line")
)

type RequestState int

const (
	Initialized RequestState = iota
	Error
	Done
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     map[string]string
	Body        []byte
	state       RequestState
}

func newRequest() *Request {
	return &Request{
		state: Initialized,
	}
}

func (r *Request) parse(p []byte) (int, error) {
	idx := bytes.Index(p, ls)
	if idx == -1 {
		// not enough data for parsing
		return 0, nil
	}

	rl, err := parseRequestLine(p[:idx])
	if err != nil {
		r.state = Error
		return 0, err
	}

	r.state = Done
	r.RequestLine = *rl
	return idx, nil
}

func (r *Request) done() bool {
	return r.state == Done || r.state == Error
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	// TODO: overflow 1024 bytes
	buf := make([]byte, 1024)
	bufi := 0

	// buf[:bufi] denote the available buffer to be parsed by request
	for !req.done() {
		// read a chunk
		n, err := reader.Read(buf[bufi:])
		// TODO: io.EOF??
		if err != nil {
			return nil, err
		}

		bufi += n
		readN, err := req.parse(buf[:bufi])
		if err != nil {
			return nil, err
		}

		// shift available buffer to left
		copy(buf, buf[:readN])
		bufi -= readN
	}

	return req, nil
}

// HTTP-version  = HTTP-name "/" DIGIT "." DIGIT
// HTTP-name     = %s"HTTP"
// request-line  = method SP request-target SP HTTP-version
func parseRequestLine(data []byte) (*RequestLine, error) {
	parts := bytes.Split(data, []byte(" "))
	if len(parts) != 3 {
		return nil, ErrMalformedRequestLine
	}

	method := parts[0]
	if !slices.Equal(method, bytes.ToUpper(method)) {
		return nil, ErrMalformedRequestLine
	}

	requestTarget := parts[1]
	if len(requestTarget) == 0 || bytes.Index(requestTarget, []byte("/")) != 0 {
		return nil, ErrMalformedRequestLine
	}

	httpVersion := parts[2]
	idx := bytes.Index(httpVersion, []byte("/"))
	version := string(httpVersion[idx+1:])
	if string(httpVersion[:idx]) != "HTTP" || version != "1.1" {
		return nil, ErrMalformedRequestLine
	}

	rl := &RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(requestTarget),
		Method:        string(method),
	}
	return rl, nil
}
