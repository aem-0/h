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
	values, exists := headers.Get("HOST")
	require.True(t, exists)
	require.Len(t, values, 1)
	assert.Equal(t, "localhost:42069", values[0])
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
	values, exists = headers.Get("HOST")
	require.True(t, exists)
	assert.Equal(t, []string{"localhost:42069", "localhost:42069"}, values)
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

func TestHeadersSet(t *testing.T) {
	tests := []struct {
		name           string
		initialValues  []string
		newValue       string
		expectedValues []string
	}{
		{
			name:           "sets new header",
			initialValues:  nil,
			newValue:       "localhost",
			expectedValues: []string{"localhost"},
		},
		{
			name:           "replaces existing value",
			initialValues:  []string{"localhost"},
			newValue:       "example.com",
			expectedValues: []string{"example.com"},
		},
		{
			name:           "replaces multiple existing values",
			initialValues:  []string{"localhost", "example.com"},
			newValue:       "api.example.com",
			expectedValues: []string{"api.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeaders()

			for _, value := range tt.initialValues {
				h.Add("Host", value)
			}

			h.Set("Host", tt.newValue)

			values, exists := h.Get("Host")

			require.True(t, exists)
			assert.Equal(t, tt.expectedValues, values)
		})
	}
}

func TestHeadersAdd(t *testing.T) {
	tests := []struct {
		name           string
		values         []string
		expectedValues []string
	}{
		{
			name:           "adds first value",
			values:         []string{"localhost"},
			expectedValues: []string{"localhost"},
		},
		{
			name:           "adds multiple values",
			values:         []string{"localhost", "example.com"},
			expectedValues: []string{"localhost", "example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeaders()

			for _, value := range tt.values {
				h.Add("Host", value)
			}

			values, exists := h.Get("Host")

			require.True(t, exists)
			assert.Equal(t, tt.expectedValues, values)
		})
	}
}
func TestHeadersDelete(t *testing.T) {
	tests := []struct {
		name       string
		headerName string
		deleteName string
	}{
		{
			name:       "deletes existing header",
			headerName: "Host",
			deleteName: "Host",
		},
		{
			name:       "delete is case insensitive",
			headerName: "Content-Type",
			deleteName: "CONTENT-TYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeaders()

			h.Set(tt.headerName, "some-value")
			h.Delete(tt.deleteName)

			_, exists := h.Get(tt.headerName)

			assert.False(t, exists)
		})
	}
}
