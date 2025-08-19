package request

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/phungducminh/httpfromtcp/internal/headers"
)

var ls = []byte("\r\n")

var (
	ErrMalformedRequest        = fmt.Errorf("request: malformed request")
	ErrMalformedRequestLine    = fmt.Errorf("request: malformed request line")
	ErrMalformedRequestHeaders = fmt.Errorf("request: malformed request headers")
)

type RequestState string

const (
	Initialized        RequestState = "Initialized"
	Error              RequestState = "Error"
	Done               RequestState = "Done"
	ParsingRequestLine RequestState = "ParsingRequestLine"
	ParsingHeaders     RequestState = "ParsingHeaders"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       RequestState
	// readN       int
}

func newRequest() *Request {
	return &Request{
		state: Initialized,
		Headers: headers.NewHeaders(),
	}
}

// parse return number of bytes read and error
// the number of bytes read will be used for moving buffer
func (r *Request) parse(p []byte) (int, error) {
	readN := 0
outer:
	for {
		switch r.state {
		case Done:
			break outer
		case Error:
			break outer
		case Initialized:
			r.state = ParsingRequestLine
		case ParsingHeaders:
			s := p[readN:]
			n, done, err := r.Headers.Parse(s)
			if err != nil {
				r.state = Error
				return readN, err
			}
			if !done {
				break outer
			}

			r.state = Done
			readN += n			
			break outer
		case ParsingRequestLine:
			rl, n, err := parseRequestLine(p)
			if err != nil {
				r.state = Error
				return 0, err
			}
			if rl == nil {
				break outer
			}

			r.RequestLine = *rl
			r.state = ParsingHeaders
			readN += n
		default:
			panic(fmt.Sprintf("unexpected request.RequestState: %#v", r.state))
		}
	}

	return readN, nil
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
		// read a chunk
		n, err := reader.Read(buf[end:])
		// TODO: io.EOF??
		if err != nil {
			return nil, err
		}

		end += n
		readN, err := req.parse(buf[:end])
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
