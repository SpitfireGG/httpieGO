package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"spitfiregg.httpFromScratch.httpieee/internal/headers"
)

type Parserstate string

const (
	BUF_SIZE int = 1024

	RequestStateInit           Parserstate = "Init"
	RequestStateDone           Parserstate = "Done"
	RequestBody                Parserstate = "Body"
	requestStateParsingHeaders Parserstate = "Header"
	RequestStateError          Parserstate = "Error"
)

var (
	clrf                       = "\r\n"
	MALFORMED_START_LINE       = fmt.Errorf("malformed start line")
	MISCONFIGURED_HTTP_VERSION = fmt.Errorf("malformed/issue in the HTTP version")
)

type Request struct {
	RequestLine   RequestLine
	Header        headers.Headers
	Body          []byte
	BodyBytesRead int
	ContentLength int
	state         Parserstate
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func GetCL(headers headers.Headers, name string, defaultVal int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal

	}
	return value
}

func ParseRequestLine(str []byte) (*RequestLine, int, error) {

	idx := bytes.Index(str, []byte(clrf))
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := str[:idx]

	parts := strings.Split(string(startLine), " ")
	if len(parts) < 3 {
		return nil, 0, MALFORMED_START_LINE
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, 0, MISCONFIGURED_HTTP_VERSION
	}

	r1 := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpParts[1],
	}

	return r1, idx + len(clrf), nil
}

func (r *Request) err() bool {
	return r.state == RequestStateError
}

func (r *Request) done() bool {
	return r.state == RequestStateDone
}

func (r *Request) Parse(data []byte) (int, error) {

	read := 0
outer:
	for {
		// 		request Line			Header field I				Header field II			HfIII		markend
		// ------|-----------------|------------------------|--------------------------|--------------|----|
		// data:"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",

		currentData := data[read:]
		slog.Info("Request#parse loop", "state", r.state, "currentData", string(currentData), "consumed", len(string(currentData)))
		if len(currentData) == 0 {
			slog.Info("no more data, breaking...")
			break outer
		}

		switch r.state {

		case RequestStateError:
			r.err()

		case RequestStateInit:
			requestLine, byteCons, err := ParseRequestLine(currentData)
			slog.Info("Bytes consumed# ", "bc", byteCons, "state", r.state)

			if err != nil {
				r.state = RequestStateError
				return read, err
			}
			if byteCons == 0 {
				slog.Info("need more data to be parsed")
				break outer
			}
			r.RequestLine = *requestLine
			read += byteCons
			slog.Info("bytesRead#", "read", read)

			slog.Info("request#state done, transitioning to header parser state")
			r.state = requestStateParsingHeaders
			continue outer

		case RequestStateDone:
			slog.Info("parsing header state")
			break outer

		case RequestBody:
			bytesNeeded := r.ContentLength - r.BodyBytesRead
			slog.Info("#bytesNeeded => ", "[%d]\n", bytesNeeded)

			if len(currentData) >= bytesNeeded {
				r.Body = append(r.Body, currentData[:bytesNeeded]...)
				r.BodyBytesRead += bytesNeeded
				read += bytesNeeded

				r.state = RequestStateDone
				break outer
			}

			r.Body = append(r.Body, currentData...)
			r.BodyBytesRead += len(currentData)
			read += len(currentData)
			break outer

		case requestStateParsingHeaders:
			n, done, err := r.Header.Parse(currentData)

			slog.Info("Read bytes", "requestStateParsingHeaders", read)

			if err != nil {
				r.state = RequestStateError
				return read, err
			}
			if n == 0 {
				break outer
			}
			read += n

			if !done {
				slog.Info("Need more data for headers.")
				break outer
			}
			slog.Info("Header parsing complete. Transitioning to check for Body.", "headerBytes", n)

			// Check for Content-Length header to see if we expect a body
			cl := GetCL(r.Header, "Content-Length", 0)
			r.ContentLength = cl

			if cl > 0 {
				r.state = RequestBody
				// We need to initialize the Body slice if it's nil
				if r.Body == nil {
					r.Body = make([]byte, 0, cl)
				}
			} else {
				r.state = RequestStateDone
			}
			continue outer

		default:
			return 0, errors.New("error: unknown error")
		}
	}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a

	}
	slogHandlerOption := slog.HandlerOptions{
		ReplaceAttr: replaceAttr,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slogHandlerOption))
	slog.SetDefault(logger)

	request := &Request{
		Header: *headers.NewHeaders(),
		state:  RequestStateInit,
	}

	buf := make([]byte, BUF_SIZE)
	bufIdx := 0

	for request.state != RequestStateDone && request.state != RequestStateError {
		slog.Info("buffer", "buffer index", bufIdx)
		if bufIdx >= len(buf) {
			// increase the buffer size
			newbuf := make([]byte, BUF_SIZE*2)
			copy(newbuf, buf[:bufIdx])
			buf = newbuf
		}
		n, err := reader.Read(buf[bufIdx:])
		slog.Info("buffer#", "buffer read", n)

		if err == io.EOF {
			if request.done() {
				slog.Debug("error buffer context", "error#", err)
				break
			} else {
				return nil, errors.New("unexpected, reached the end of the file")
			}
		}
		if err != nil {
			request.state = RequestStateError
			return nil, fmt.Errorf("error reading data: %w", err)
		}
		bufIdx += n
		bytesCons, err := request.Parse(buf[:bufIdx])
		if err != nil {
			request.state = RequestStateError
			return nil, fmt.Errorf("error parsing the buffer, %w", err)
		}
		if bytesCons > 0 {
			copy(buf, buf[bytesCons:bufIdx])
			bufIdx -= bytesCons
		}
		if request.done() {
			break
		}
	}
	return request, nil

}
