package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	value, exists := headers.Get("HOST")
	assert.True(t, exists)
	assert.Equal(t, "localhost:42069", value)
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing he	ader
	headers = NewHeaders()
	data = []byte("        Host : localhost:42069        \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Bad header name
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	value, exists = headers.Get("HOST")
	assert.True(t, exists)
	assert.Equal(t, "localhost:42069,localhost:42069", value)
	assert.False(t, done)
}

func TestIsToken(t *testing.T) {
	assert.True(t, isToken([]byte("Host")))
	assert.True(t, isToken([]byte("Content-Type")))
	assert.True(t, isToken([]byte("X_Custom_Header")))

	assert.False(t, isToken([]byte("")))
	assert.False(t, isToken([]byte("Content Type")))
	assert.False(t, isToken([]byte("Content:Type")))
	assert.False(t, isToken([]byte("Content@Type")))
}

func TestParseHeaderFieldName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		Err   bool
	}{
		{
			name:  "valid header",
			input: "Host: localhost\r\n",
			Err:   false,
		},
		{
			name:  "space before field name",
			input: " Host: localhost\r\n",
			Err:   true,
		},
		{
			name:  "tab before field name",
			input: "\tHost: localhost\r\n",
			Err:   true,
		},
		{
			name:  "invalid character",
			input: "Ho@st: localhost\r\n",
			Err:   true,
		},
		{
			name:  "empty field name",
			input: ": localhost\r\n",
			Err:   true,
		},
		{
			name:  "space before colon",
			input: "Host : localhost\r\n",
			Err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseHeader([]byte(tt.input))
			if tt.Err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
