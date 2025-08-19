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
		expectDone    bool
	}{
		{
			data:          "Host: localhost:42069\r\nContent-Type: application/json\r\n\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{"Host": "localhost:42069", "Content-Type": "application/json"},
			expectReadLen: 57,
			expectDone:    true,
		},
		{
			description:   "spacing headers",
			data:          "       Host: localhost:42069       \r\n\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{"Host": "localhost:42069"},
			expectReadLen: 39,
			expectDone:    true,
		},
		{
			description:   "valid allowed characters",
			data:          "H!1ost: localhost:42069\r\n\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{"h!1ost": "localhost:42069"},
			expectReadLen: 27,
			expectDone:    true,
		},
		{
			data:          "H@1ost: localhost:42069\r\n\r\n",
			expectErr:     ErrMalformedHeaders,
			expectHeaders: map[string]string{},
			expectReadLen: 0,
			expectDone:    false,
		},
		{
			description:   "invalid spacing header",
			data:          "       Host : localhost:42069       \r\n\r\n",
			expectErr:     ErrMalformedHeaders,
			expectHeaders: map[string]string{},
			expectReadLen: 0,
			expectDone:    false,
		},
		{
			description:   "empty headers",
			data:          "\r\n",
			expectErr:     nil,
			expectHeaders: map[string]string{},
			expectReadLen: 2,
			expectDone:    true,
		},
		{
			description:   "missing end of header",
			data:          "Host : localhost:42069",
			expectErr:     nil,
			expectHeaders: map[string]string{},
			expectReadLen: 0,
			expectDone:    false,
		},
	}

	for _, tt := range tests {
		headers := NewHeaders()
		readLen, done, err := headers.Parse([]byte(tt.data))
		require.Equal(t, tt.expectErr, err)
		for k, v := range tt.expectHeaders {
			assert.Equal(t, v, headers.Get(k))
		}
		assert.Equal(t, tt.expectReadLen, readLen)
		assert.Equal(t, tt.expectDone, done)
	}
}

func TestHeadersParseMultipleValues(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nSet-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n")
	_, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers.Get("Set-Person"))
	assert.True(t, done)
}
