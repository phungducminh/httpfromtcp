package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            []byte
	numBytesPerRead int
	pos             int
}

func newChunkReader(data []byte, numBytesPerRead int) *chunkReader {
	return &chunkReader{
		data:            data,
		numBytesPerRead: numBytesPerRead,
		pos:             0,
	}
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
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
