package request

import (
	"io"
	"testing"

	"github.com/phungducminh/httpfromtcp/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            []byte
	numBytesPerRead int
	end             int
}

func newChunkReader(data []byte, numBytesPerRead int) *chunkReader {
	return &chunkReader{
		data:            data,
		numBytesPerRead: numBytesPerRead,
		end:             0, // last read position
	}
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.end >= len(cr.data) {
		return 0, io.EOF
	}
	nexti := min(cr.end+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.end:nexti])
	cr.end += n
	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	r, err := RequestFromReader(newChunkReader([]byte("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 3))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(newChunkReader([]byte("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 1))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid method
	r, err = RequestFromReader(newChunkReader([]byte("get / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 5))
	require.Error(t, err)
	assert.Equal(t, err, ErrMalformedRequestLine)

	// Test: Invalid version
	r, err = RequestFromReader(newChunkReader([]byte("get / HTTP/1.2\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 20))
	require.Error(t, err)
	assert.Equal(t, err, ErrMalformedRequestLine)

	// Test: Out of order
	r, err = RequestFromReader(newChunkReader([]byte("GET HTTP/1.1 /\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 5))
	require.Error(t, err)
	assert.Equal(t, err, ErrMalformedRequestLine)

	// Test: Invalid http version
	r, err = RequestFromReader(newChunkReader([]byte("GET / TCP/1.1 \r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 10))
	require.Error(t, err)
	assert.Equal(t, err, ErrMalformedRequestLine)

	// Test: Lacking method
	_, err = RequestFromReader(newChunkReader([]byte("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"), 10))
	require.Error(t, err)
	assert.Equal(t, err, ErrMalformedRequestLine)
}

func TestRequestFromReaderParseHeaders(t *testing.T) {
	// Test: Standard Headers
	tests := []struct {
		data            string
		numBytesPerRead int
		expectHeaders   map[string]string
		expectErr       error
	}{
		{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 100,
			expectHeaders:   map[string]string{"host": "localhost:42069", "user-agent": "curl/7.81.0", "accept": "*/*"},
			expectErr:       nil,
		},
		{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
			expectHeaders:   map[string]string{"host": "localhost:42069", "user-agent": "curl/7.81.0", "accept": "*/*"},
			expectErr:       nil,
		},
		{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nBar: abc\r\nbar:xyz\r\n\r\n",
			numBytesPerRead: 3,
			expectHeaders:   map[string]string{"host": "localhost:42069", "bar": "abc, xyz"},
			expectErr:       nil,
		},
		{
			data:            "GET / HTTP/1.1\r\n\r\n",
			numBytesPerRead: 3,
			expectHeaders:   map[string]string{},
			expectErr:       nil,
		},
		{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
			expectHeaders:   map[string]string{},
			expectErr:       headers.ErrMalformedHeaders,
		},
	}

	for _, tt := range tests {
		r := newChunkReader([]byte(tt.data), tt.numBytesPerRead)
		req, err := RequestFromReader(r)
		if tt.expectErr != nil && tt.expectErr != err {
			t.Errorf("data='%s', numBytesPerRead=%d, expect error=%v, actual=%v", tt.data, tt.numBytesPerRead, tt.expectErr, err)
		}
		if tt.expectErr == nil {
			require.NotNil(t, req, tt.data)
			for k, v := range tt.expectHeaders {
				actual := req.Headers.Get(k)
				if v != actual {
					t.Errorf("data='%s', numBytesPerRead=%d, key=%s, expect value=%s, actual value=%s", tt.data, tt.numBytesPerRead, k, v, actual)
				}
			}
		}
	}
}

func TestRequestFromReaderParseBody(t *testing.T) {
	tests := []struct {
		description     string
		data            string
		numBytesPerRead int
		expecteError    error
		expectBody      string
	}{
		{
			description: "happy case",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
			expecteError:    nil,
			expectBody:      "hello world!\n",
		},
		{
			description: "content-length header's value and body length mismatch",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			numBytesPerRead: 3,
			expecteError:    ErrMalformedRequestBody,
			expectBody:      "",
		},
		{
			description: "content-length header's value is 0 and empty body",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: 3,
			expecteError:    nil,
			expectBody:      "",
		},
		{
			description: "no content-length header and empty body",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 3,
			expecteError:    nil,
			expectBody:      "",
		},
		{
			description: "no content-length header and body exist",
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
			expecteError:    nil,
			expectBody:      "",
		},
	}

	for _, tt := range tests {
		reader := newChunkReader([]byte(tt.data), tt.numBytesPerRead)
		r, err := RequestFromReader(reader)
		if tt.expecteError != nil {
			require.Error(t, err, tt.description)
			assert.Equal(t, tt.expecteError, err, tt.description)
		} else {
			require.NotNil(t, r, tt.description)
			require.NotNil(t, r.Body, tt.description)
			assert.Equal(t, tt.expectBody, string(r.Body), tt.description)
		}
	}
}
