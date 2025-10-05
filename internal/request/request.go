package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

type status int

const (
	initialized = iota
	done
)

type Request struct {
	status      status
	RequestLine Line
}

type Line struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func FromReader(reader io.Reader) (*Request, error) {
	request := &Request{status: initialized}
	buffer := make([]byte, 8)
	bytesRead := 0

	for request.status != done {
		if len(buffer) == cap(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		readCount, err := reader.Read(buffer[bytesRead:])

		if err != nil {
			if err == io.EOF {
				request.status = done
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
			buffer = buffer[parsedCount:]
			bytesRead -= parsedCount
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.status == done {
		return 0, errors.New("request already parsed")
	}

	if r.status != initialized {
		return 0, fmt.Errorf("unknown state: %d", r.status)
	}

	parsedCount, requestLine, err := parseRequestLine(data)

	if parsedCount == 0 || err != nil {
		return parsedCount, err
	}

	r.RequestLine = *requestLine
	r.status = done
	return parsedCount, nil
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
