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

func Parse(data []byte) (*Headers, int, error) {
	h := NewHeaders()
	// no header case
	if bytes.Index(data, fieldLineDelimiter) == 0 {
		return h, len(fieldLineDelimiter), nil
	}

	endi := bytes.Index(data, headersDelimiter)
	if endi == -1 {
		return nil, 0, nil
	}
	n := 0
	for {
		linei := bytes.Index(data[n:], fieldLineDelimiter)
		if linei == -1 {
			return nil, 0, nil
		}

		if linei == 0 {
			// end of headers
			break
		}

		// idx: the index starting from n -> need to take slice [n:n+idx]
		buf := data[n : n+linei]
		coloni := bytes.Index(buf, []byte(":"))
		if coloni == -1 || (len(buf) >= coloni && buf[coloni-1] == ' ') {
			return nil, 0, ErrMalformedHeaders
		}

		fieldName := strings.TrimSpace(string(buf[:coloni]))
		fieldValue := strings.TrimSpace(string(buf[coloni+1:]))
		if !isToken(fieldName) {
			return nil, 0, ErrMalformedHeaders
		}

		val := h.Get(fieldName)
		if val != "" {
			h.Set(fieldName, val+", "+fieldValue)
		} else {
			h.Set(fieldName, fieldValue)
		}

		n += linei + len(fieldLineDelimiter)
	}

	return h, endi + len(headersDelimiter), nil
}
