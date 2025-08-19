package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var ls = []byte("\r\n")

type Headers struct {
	m map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		m: make(map[string]string),
	}
}

func (h *Headers) Get(key string) string {
	return h.m[key]
}

func (h *Headers) Set(key, value string) {
	h.m[key] = value
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	n := 0
	for {
		idx := bytes.Index(data[n:], ls)
		if idx == -1 {
			return 0, false, nil
		}

		if idx == 0 {
			// end of headers
			break
		}

		// idx: the index starting from n -> need to take slice [n:n+idx]
		buf := data[n : n+idx]
		colonIdx := bytes.Index(buf, []byte(":"))
		if colonIdx == -1 || (len(buf) >= colonIdx && buf[colonIdx-1] == ' ') {
			return 0, false, fmt.Errorf("headers: malformed header name")
		}

		fieldName := strings.TrimSpace(string(buf[:colonIdx]))
		fieldValue := strings.TrimSpace(string(buf[colonIdx+1:]))
		h.Set(fieldName, fieldValue)

		n += idx + len(ls)
	}

	return n, true, nil
}
