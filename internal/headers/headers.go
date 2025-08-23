package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var fieldLineDelimiter = []byte("\r\n")
var headersDelimiter = []byte("\r\n\r\n")

var ErrMalformedHeaders = fmt.Errorf("headers: malformed headers")

type Headers struct {
	kv map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		kv: make(map[string]string),
	}
}

func (h *Headers) Get(key string) string {
	return h.kv[strings.ToLower(key)]
}

func (h *Headers) Set(key, value string) {
	h.kv[strings.ToLower(key)] = value
}

func (h *Headers) Len() int {
	return len(h.kv)
}

func (h *Headers) ForEach(fn func(string, string)) {
	for k, v := range h.kv {
		fn(k, v)
	}
}

func isToken(key string) bool {
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c >= 'a' && c <= 'z' {
			continue
		} else if c >= 'A' && c <= 'Z' {
			continue
		} else if c >= '0' && c <= '9' {
			continue
		} else if c == '!' || c == '#' || c == '$' || c == '%' || c == '&' || c == '\'' || c == '*' || c == '+' || c == '-' || c == '.' || c == '^' || c == '_' || c == '`' || c == '|' || c == '~' {
			continue
		} else {
			return false
		}
	}

	return true
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	// no header case
	if bytes.Index(data, fieldLineDelimiter) == 0 {
		return len(fieldLineDelimiter), true, nil
	}

	ei := bytes.Index(data, headersDelimiter)
	if ei == -1 {
		return 0, false, nil
	}
	n := 0
	for {
		idx := bytes.Index(data[n:], fieldLineDelimiter)
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
			return 0, false, ErrMalformedHeaders
		}

		fieldName := strings.TrimSpace(string(buf[:colonIdx]))
		fieldValue := strings.TrimSpace(string(buf[colonIdx+1:]))
		if !isToken(fieldName) {
			return 0, false, ErrMalformedHeaders
		}

		val := h.Get(fieldName)
		if val != "" {
			h.Set(fieldName, val+", "+fieldValue)
		} else {
			h.Set(fieldName, fieldValue)
		}

		n += idx + len(fieldLineDelimiter)
	}

	return ei + len(headersDelimiter), true, nil
}
