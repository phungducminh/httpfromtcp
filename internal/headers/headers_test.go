package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	tests := []struct {
		description   string
		data          string
		expectErr     error
		expectHeaders map[string]string
		expectReadLen int
	}{
		{
			data:          "Host: localhost:42069\r\nContent-Type: application/json\r\n\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{"Host": "localhost:42069", "Content-Type": "application/json"},
			expectReadLen: 57,
		},
		{
			description:   "spacing headers",
			data:          "       Host: localhost:42069       \r\n\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{"Host": "localhost:42069"},
			expectReadLen: 39,
		},
		{
			description:   "valid allowed characters",
			data:          "H!1ost: localhost:42069\r\n\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{"h!1ost": "localhost:42069"},
			expectReadLen: 27,
		},
		{
			data:          "H@1ost: localhost:42069\r\n\r\n",
			expectErr:     ErrMalformedHeaders,
			expectHeaders: map[string]string{},
			expectReadLen: 0,
		},
		{
			description:   "invalid spacing header",
			data:          "       Host : localhost:42069       \r\n\r\n",
			expectErr:     ErrMalformedHeaders,
			expectHeaders: map[string]string{},
			expectReadLen: 0,
		},
		{
			description:   "empty headers",
			data:          "\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{},
			expectReadLen: 2,
		},
		{
			description:   "missing end of header",
			data:          "Host : localhost:42069",
			expectErr:     nil,
			expectHeaders: map[string]string{},
			expectReadLen: 0,
		},
	}

	for _, tt := range tests {
		h, n, err := Parse([]byte(tt.data), false)
		require.Equal(t, tt.expectErr, err)
		for k, v := range tt.expectHeaders {
			assert.Equal(t, v, h.Get(k))
		}
		assert.Equal(t, tt.expectReadLen, n)
		if n != 0 {
			require.NotNil(t, h)
		}
	}
}

func TestHeadersParseMultipleValues(t *testing.T) {
	data := []byte("Host: localhost:42069\r\nSet-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n")
	h, n, err := Parse(data, false)
	require.NoError(t, err)
	require.NotNil(t, h)
	assert.Equal(t, "localhost:42069", h.Get("Host"))
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", h.Get("Set-Person"))
	assert.Equal(t, n, 109)
}
