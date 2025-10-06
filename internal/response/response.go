package response

import (
	"errors"
	"fmt"
	"io"

	"github.com/MadhurSahu/tcp-to-http/internal/headers"
)

type StatusCode int

const (
	StatusCodeOK                  = 200
	StatusCodeBadRequest          = 400
	StatusCodeInternalServerError = 500
)

type WriteStatus int

const (
	WriteStatusLine = iota
	WriteStatusHeaders
	WriteStatusBody
)

const (
	bodyBadRequest = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
	bodyInternalServerError = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
)

type Writer struct {
	status WriteStatus
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		status: WriteStatusLine,
	}
}

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.status != WriteStatusBody {
		return 0, errors.New("cannot write body yet (or has already been written)")
	}

	_, err := w.writer.Write(data)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func (w *Writer) WriteError(code StatusCode) error {
	body := bodyBadRequest
	if code == StatusCodeInternalServerError {
		body = bodyInternalServerError
	}

	errorHeaders := headers.GetDefaultHeaders(len(body))
	errorHeaders.Overwrite("Content-Type", "text/html")

	err := w.WriteStatusLine(code)
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}

	err = w.WriteHeaders(errorHeaders)
	if err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	_, err = w.WriteBody([]byte(body))
	if err != nil {
		return fmt.Errorf("error writing response body: %w", err)
	}

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.status != WriteStatusHeaders {
		return errors.New("cannot write headers yet (or has already been written)")
	}

	for key, val := range headers {
		_, err := w.writer.Write([]byte(key + ": " + val + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	w.status = WriteStatusBody
	return err
}

func (w *Writer) WriteStatusLine(code StatusCode) error {
	if w.status != WriteStatusLine {
		return errors.New("cannot write status line twice")
	}

	str := ""
	switch code {
	case StatusCodeOK:
		str = "HTTP/1.1 200 OK"
	case StatusCodeBadRequest:
		str = "HTTP/1.1 400 Bad Request"
	case StatusCodeInternalServerError:
		str = "HTTP/1.1 500 Internal Server Error"
	default:
		str = fmt.Sprintf("HTTP/1.1 %d ", code)
	}

	_, err := w.writer.Write([]byte(str + "\r\n"))
	w.status = WriteStatusHeaders
	return err
}
