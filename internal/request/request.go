package request

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"slices"
)

var ls = []byte("\r\n")

var (
	ErrMalformedRequest     = fmt.Errorf("request: malformed request")
	ErrMalformedRequestLine = fmt.Errorf("request: malformed request line")
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
}

// HTTP-version  = HTTP-name "/" DIGIT "." DIGIT
// HTTP-name     = %s"HTTP"
// request-line  = method SP request-target SP HTTP-version
func parseRequestLine(data []byte) (RequestLine, error) {
	parts := bytes.Split(data, []byte(" "))
	if len(parts) != 3 {
		return RequestLine{}, ErrMalformedRequestLine
	}

	method := parts[0]
	if !slices.Equal(method, bytes.ToUpper(method)) {
		return RequestLine{}, ErrMalformedRequestLine
	}

	requestTarget := parts[1]
	if len(requestTarget) == 0 || bytes.Index(requestTarget, []byte("/")) != 0 {
		return RequestLine{}, ErrMalformedRequest
	}

	httpVersion := parts[2]
	idx := bytes.Index(httpVersion, []byte("/"))
	version := string(httpVersion[idx+1:])
	if string(httpVersion[:idx]) != "HTTP" || version != "1.1" {
		return RequestLine{}, ErrMalformedRequest
	}

	rl := RequestLine{
		HttpVersion:   string(version),
		RequestTarget: string(requestTarget),
		Method:        string(method),
	}
	return rl, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("request: failed to read all, err=%v", err)
	}
	lines := bytes.Split(data, ls)
	if len(lines) <= 0 {
		return nil, ErrMalformedRequest
	}

	rl, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	req := &Request{RequestLine: rl}
	return req, nil
}
