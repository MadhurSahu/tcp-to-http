package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("host:localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 22, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("  host: localhost:42069  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 27, n)
	assert.False(t, done)

	// Test: Valid multiple headers
	headers = NewHeaders()
	headers["host"] = "localhost:42069"
	data = []byte("host:localhost:8080\r\nClient:Testify\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069, localhost:8080", headers["host"])
	assert.Equal(t, 21, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n{\"hello\":\"hi\"}")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	//Test: Lower case handling
	headers = NewHeaders()
	data = []byte("Host:localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 22, n)
	assert.False(t, done)

	//Test: Invalid header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
