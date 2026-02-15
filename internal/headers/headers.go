package headers

import (
	"bytes"
	"fmt"
	"log/slog"

	"strings"
)

/*
'Host: localhost:42069'
'          Host: localhost:42069    '
*/

type Issue string

const (
	MalformedHeaderIssue Issue = "Malformed header, try checking the header explictly"
)

type Headers struct {
	Header map[string]string
}

var clrf = []byte("\r\n")

func NewHeaders() *Headers {
	/* 	return map[string]string{} */
	headers := &Headers{
		Header: make(map[string]string),
	}
	return headers
}

func (h *Headers) Set(key, value string) {
	h.Header[strings.ToLower(key)] = value
}

func (h Headers) Add(key, value string) {
	lowerKey := strings.ToLower(key)
	if existing, exists := h.Header[lowerKey]; exists {
		h.Header[lowerKey] = fmt.Sprintf("%s,%s", existing, value)
	} else {
		h.Header[lowerKey] = value
	}
}

func (h *Headers) Get(key string) (string, bool) {
	str, ok := (h.Header[strings.ToLower(string(key))])
	if !ok {
		return "", false
	}
	return str, true
}

func isToken(ch byte) bool {
	switch {
	case ch >= 'a' && ch <= 'z':
		return true
	case ch >= '0' && ch <= '9':
		return true
	case ch >= 'A' && ch <= 'Z':
		return true
	case ch == '!' || ch == '#' || ch == '$' || ch == '%' || ch == '&' ||
		ch == '\'' || ch == '*' || ch == '+' || ch == '-' || ch == '.' ||
		ch == '^' || ch == '_' || ch == '`' || ch == '|' || ch == '~':
		return true
	default:
		return false
	}
}

func ParseHeaders(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2) // split into host and address
	slog.Info("ParseHeaders", "fieldLine", string(fieldLine))
	if len(parts) != 2 {
		return "", "", fmt.Errorf(string(MalformedHeaderIssue))
	}

	if bytes.HasPrefix(fieldLine, []byte(" ")) || bytes.HasPrefix(fieldLine, []byte("\t")) {
		return "", "", fmt.Errorf(string(MalformedHeaderIssue))
	}

	key := bytes.TrimSpace(parts[0])
	value := bytes.TrimSpace(parts[1])

	if len(key) == 0 {
		return "", "", fmt.Errorf(string(MalformedHeaderIssue))
	}

	return string(key), string(value), nil
}

func (h *Headers) Parse(data []byte) (int, bool, error) {

	/* 		"Host: localhost:42069\r\n\r\n" */

	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], clrf) // do the subslicing
		if idx == -1 {
			break
		}
		if idx == 0 { // found \r\n immediately — end of headers
			done = true
			read += len(clrf) // consume the empty line's \r\n
			break
		}

		/* 		fmt.Printf("header: \"%s\"\n", string(data[read:read+idx])) */

		key, value, err := ParseHeaders(data[read : read+idx]) // reaches till the defined clrf|
		for _, ch := range key {
			if !isToken(byte(ch)) {
				return 0, true, fmt.Errorf(string(MalformedHeaderIssue))
			}
		}

		if err != nil {
			done = false
			return 0, done, fmt.Errorf(string(MalformedHeaderIssue))
		}
		h.Add(key, value)
		read += idx + len(clrf)
		slog.Info("headers.go", "read", read)
	}
	return read, done, nil
}
