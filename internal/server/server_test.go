package server

import (
	"bytes"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/aem-0/h/internal/request"
	"github.com/aem-0/h/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConn struct {
	io.Reader
	io.Writer
	closed bool
}

func (c *testConn) Close() error {
	c.closed = true
	return nil
}

func TestRunConnection(t *testing.T) {
	requestData := "GET / HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"\r\n"

	var responseBuffer bytes.Buffer

	conn := &testConn{
		Reader: bytes.NewBufferString(requestData),
		Writer: &responseBuffer,
	}

	handlerCalled := false

	handler := func(w *response.Writer, req *request.Request) {
		handlerCalled = true

		assert.Equal(t, "GET", req.RequestLine.Method)
		assert.Equal(t, "/", req.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", req.RequestLine.HttpVersion)

		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(*response.GetDefaultHeaders(0))
	}

	s := &Server{
		handler: handler,
	}

	runConnection(s, conn)

	assert.True(t, handlerCalled)
	assert.True(t, conn.closed)
	assert.Contains(t, responseBuffer.String(), "HTTP/1.1 200 OK\r\n")
}

func TestRunConnectionBadRequest(t *testing.T) {
	requestData := "GET / HTTP/1.1\r\n" +
		"Host localhost:42069\r\n" +
		"\r\n"

	var responseBuffer bytes.Buffer

	conn := &testConn{
		Reader: bytes.NewBufferString(requestData),
		Writer: &responseBuffer,
	}

	handlerCalled := false

	handler := func(w *response.Writer, req *request.Request) {
		handlerCalled = true
	}

	s := &Server{
		handler: handler,
	}

	runConnection(s, conn)

	assert.False(t, handlerCalled)
	assert.True(t, conn.closed)
	assert.Contains(t, responseBuffer.String(), "HTTP/1.1 400 Bad Request\r\n")
}

func TestServe(t *testing.T) {
	handler := func(w *response.Writer, req *request.Request) {
		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(*response.GetDefaultHeaders(5))
		w.WriteBody([]byte("hello"))
	}

	server, err := Serve(0, handler)
	require.NoError(t, err)
	require.NotNil(t, server)

	conn, err := net.Dial("tcp", server.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()
	defer server.Close()
	_, err = conn.Write([]byte(
		"GET / HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"\r\n",
	))
	require.NoError(t, err)

	responseBytes, err := io.ReadAll(conn)
	require.NoError(t, err)

	response := string(responseBytes)

	assert.Contains(t, response, "HTTP/1.1 200 OK\r\n")
	assert.Contains(t, strings.ToLower(response), "content-length: 5\r\n")
	assert.Contains(t, response, "hello")
}
