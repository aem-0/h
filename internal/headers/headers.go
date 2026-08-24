package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
	if len(str) == 0 {
		return false
	}
	for _, ch := range str {
		switch {
		case ch >= 'A' && ch <= 'Z', ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9':
			continue
		case ch == '#', ch == '$', ch == '%', ch == '&', ch == '\'', ch == '*',
			ch == '+', ch == '-', ch == '.', ch == '^', ch == '_', ch == '`',
			ch == '|', ch == '~':
			continue
		default:
			return false
		}

	}
	return true
}

var rn = []byte("\r\n")

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("bad field line")
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if !isToken(name) {
		return "", "", fmt.Errorf("bad field name")
	}

	return string(name), string(value), nil
}

type Headers struct {
	headers map[string][]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string][]string{},
	}
}

func (h *Headers) Get(name string) ([]string, bool) {
	values, ok := h.headers[strings.ToLower(name)]
	return values, ok
}

func (h *Headers) Add(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = append(h.headers[name], value)
}

func (h *Headers) Delete(name string) {
	name = strings.ToLower(name)
	delete(h.headers, name)
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = []string{value}
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for n, values := range h.headers {
		for _, value := range values {
			cb(n, value)
		}
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	from := 0
	done := false

	for {
		to := bytes.Index(data[from:], rn)
		fmt.Printf(" (%d) - %d\n", from, to)
		if to == -1 {
			break
		}

		if to == 0 {
			done = true
			from += len(rn)
			break
		}
		fmt.Printf("header: \"%s\"\n", string(data[from:from+to]))
		name, value, err := parseHeader(data[from : from+to])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("bad header name")
		}

		from += to + len(rn)
		h.Add(name, value)
	}
	return from, done, nil
}
