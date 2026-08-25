package response

import (
	"bytes"
	"testing"

	"github.com/aem-0/h/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteStatusLine(t *testing.T) {
	tests := []struct {
		name       string
		statusCode StatusCode
		expected   string
	}{
		{
			name:       "OK",
			statusCode: StatusOk,
			expected:   "HTTP/1.1 200 OK\r\n",
		},
		{
			name:       "Bad Request",
			statusCode: StatusBadRequest,
			expected:   "HTTP/1.1 400 Bad Request\r\n",
		},
		{
			name:       "Internal Server Error",
			statusCode: StatusInternalServerError,
			expected:   "HTTP/1.1 500 Internal Server Error\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)

			err := writer.WriteStatusLine(tt.statusCode)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestWriteStatusLineInvalidStatus(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	err := writer.WriteStatusLine(StatusCode(999))

	require.Error(t, err)
	assert.Empty(t, buf.String())
}

func TestWriteHeaders(t *testing.T) {
	h := headers.NewHeaders()
	h.Set("Content-Length", "13")
	h.Set("Content-Type", "text/plain")
	h.Set("Connection", "close")

	var buf bytes.Buffer
	writer := NewWriter(&buf)

	err := writer.WriteHeaders(*h)

	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "content-length: 13\r\n")
	assert.Contains(t, output, "content-type: text/plain\r\n")
	assert.Contains(t, output, "connection: close\r\n")
	assert.True(t, bytes.HasSuffix([]byte(output), []byte("\r\n\r\n")))
}

func TestWriteBody(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	body := []byte("Hello, world!")

	n, err := writer.WriteBody(body)

	require.NoError(t, err)
	assert.Equal(t, len(body), n)
	assert.Equal(t, "Hello, world!", buf.String())
}
