package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/MadhurSahu/tcp-to-http/internal/headers"
)

type status int

const (
	requestStatusInitialized = iota
	requestStatusParsingHeaders
	requestStatusDone
)

type Request struct {
	Headers     headers.Headers
	RequestLine Line
	status      status
}

type Line struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func FromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		Headers: headers.NewHeaders(),
		status:  requestStatusInitialized,
	}
	buffer := make([]byte, 8)
	bytesRead := 0

	for request.status != requestStatusDone {
		if bytesRead >= len(buffer) {
			println("New size:", len(buffer)*2)
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		readCount, err := reader.Read(buffer[bytesRead:])

		if err != nil {
			if err == io.EOF {
				request.status = requestStatusDone
				break
			}
			return nil, fmt.Errorf("error reading request: %w", err)
		}

		bytesRead += readCount

		parsedCount, err := request.parse(buffer[:bytesRead])
		if err != nil {
			return nil, err
		}

		if parsedCount != 0 {
			copy(buffer, buffer[parsedCount:])
			bytesRead -= parsedCount
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalParsedBytes := 0

	for r.status != requestStatusDone {
		n, err := r.parseSingle(data[totalParsedBytes:])
		totalParsedBytes += n

		if n == 0 || err != nil {
			return totalParsedBytes, err
		}
	}

	return totalParsedBytes, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.status {
	case requestStatusInitialized:
		n, requestLine, err := parseRequestLine(data)

		if n == 0 || err != nil {
			return n, err
		}

		r.RequestLine = *requestLine
		r.status = requestStatusParsingHeaders
		return n, nil
	case requestStatusParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return n, err
		}

		if done {
			r.status = requestStatusDone
		}

		return n, err

	case requestStatusDone:
		return 0, errors.New("request already parsed")
	default:
		return 0, fmt.Errorf("unknown state: %d", r.status)
	}
}

func parseRequestLine(data []byte) (int, *Line, error) {
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	str := string(data)

	if !strings.Contains(str, "\r\n") {
		return 0, nil, nil
	}

	header := strings.Split(str, "\r\n")[0]
	headerCount := len(header)
	parts := strings.Split(header, " ")

	if len(parts) != 3 {
		return 0, nil, errors.New("invalid request line")
	}

	if !slices.Contains(validMethods, parts[0]) {
		return 0, nil, fmt.Errorf("invalid method: %s", parts[0])
	}

	if !strings.HasPrefix(parts[1], "/") {
		return 0, nil, fmt.Errorf("invalid request target: %s", parts[1])
	}

	if parts[2] != "HTTP/1.1" {
		return 0, nil, errors.New("only HTTP/1.1 is supported")
	}

	version, _ := strings.CutPrefix(parts[2], "HTTP/")

	requestLine := &Line{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version,
	}

	return headerCount + 2, requestLine, nil
}
