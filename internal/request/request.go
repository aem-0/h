package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/aem-0/h/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string
	state       parserState
}

func getContentLength(h *headers.Headers) (int, bool, error) {
	values, exists := h.Get("Content-Length")
	if !exists {
		return 0, false, nil
	}

	if len(values) != 1 {
		return 0, false, fmt.Errorf("multiple Content-Length headers")
	}

	length, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, false, fmt.Errorf("invalid Content-Length")
	}

	if length < 0 {
		return 0, false, fmt.Errorf("negative Content-Length")
	}

	return length, true, nil
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: *headers.NewHeaders(),
		Body:    "",
	}
}

var ErrorBadStartLine = fmt.Errorf("bad start line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")

var SEPARATOR = []byte("\r\n")

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorBadStartLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() (bool, error) {
	length, exists, err := getContentLength(&r.Headers)
	if err != nil {
		return false, err
	}

	return exists && length > 0, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n

			if done {
				hasBody, err := r.hasBody()
				if err != nil {
					r.state = StateError
					return 0, err
				}

				if hasBody {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
		case StateBody:
			length, exists, err := getContentLength(&r.Headers)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if !exists || length == 0 {
				r.state = StateDone
				break
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining
			if len(r.Body) == length {
				r.state = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("Somehow we have programmed poorly")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil
}
