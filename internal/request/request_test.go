package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in the request line
	_, err = FromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Standard Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Headers))

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nPerson: Madhur\r\nPerson: Preeti\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "Madhur, Preeti", r.Headers["person"])

	// Test: Case insensitive headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nperson: Madhur\r\nPerson: Preeti\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "Madhur, Preeti", r.Headers["person"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = FromReader(reader)
	require.Error(t, err)
}
