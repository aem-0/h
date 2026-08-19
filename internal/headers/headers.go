package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(str []byte) bool {
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
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("bad field name")
	}

	return string(name), string(value), nil
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) (string, bool) {
	str, ok := h.headers[strings.ToLower(name)]
	return str, ok
}

func (h *Headers) Replace(name, value string) {
	name = strings.ToLower(name)
	h.headers[name] = value
}

func (h *Headers) Delete(name string) {
	name = strings.ToLower(name)
	delete(h.headers, name)
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[strings.ToLower(name)] = value
	}
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	from := 0
	done := false

	for {
		to := bytes.Index(data[from:], rn)
		fmt.Printf("parsing header (%d) - %d\n", from, to)
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
		h.Set(name, value)
	}
	return from, done, nil
}
