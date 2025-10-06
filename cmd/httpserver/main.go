package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/MadhurSahu/tcp-to-http/internal/headers"
	"github.com/MadhurSahu/tcp-to-http/internal/request"
	"github.com/MadhurSahu/tcp-to-http/internal/response"
	"github.com/MadhurSahu/tcp-to-http/internal/server"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) *server.HandlerError {
	path := req.RequestLine.RequestTarget

	internalError := &server.HandlerError{
		StatusCode: response.StatusCodeInternalServerError,
	}

	badRequestError := &server.HandlerError{
		StatusCode: response.StatusCodeBadRequest,
	}

	if path == "/myproblem" {
		return internalError
	}

	if path == "/yourproblem" {
		return badRequestError
	}

	if path == "/video" {
		err := w.WriteStatusLine(response.StatusCodeOK)
		if err != nil {
			log.Println(err)
			return internalError
		}

		file, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Println(err)
			return internalError
		}

		h := headers.GetDefaultHeaders(len(file))
		h.Overwrite("Content-Type", "video/mp4")
		err = w.WriteHeaders(h)
		if err != nil {
			log.Println(err)
			return internalError
		}

		_, err = w.WriteBody(file)
		if err != nil {
			log.Println(err)
			return internalError
		}
		return nil
	}

	if strings.HasPrefix(path, "/httpbin/") {
		endpoint := strings.TrimPrefix(path, "/httpbin/")

		if endpoint == "" {
			return badRequestError
		}

		err := w.WriteStatusLine(response.StatusCodeOK)
		if err != nil {
			return internalError
		}

		h := headers.GetDefaultHeaders(0)
		h.Delete("Content-Length")
		h.Overwrite("Transfer-Encoding", "chunked")
		h.Overwrite("Trailer", "X-Content-SHA256, X-Content-Length")

		err = w.WriteHeaders(h)
		if err != nil {
			log.Println(err)
			return internalError
		}

		//docker run -p 8080:80 kennethreitz/httpbin
		res, err := http.Get("http://localhost:8080/" + endpoint)
		if err != nil {
			log.Println(err)
			return internalError
		}
		defer res.Body.Close()

		full := make([]byte, 0)
		body := make([]byte, 1024)

		for {
			n, err := res.Body.Read(body)
			if n > 0 {
				_, err = w.WriteChunkedBody(body[:n])
				if err != nil {
					fmt.Println("Error writing chunked body:", err)
					break
				}
				full = append(full, body[:n]...)
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Println(err)
				return internalError
			}
		}

		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			log.Println(err)
			return internalError
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256(full))
		contentLength := fmt.Sprintf("%d", len(full))
		trailerHeaders := headers.NewHeaders()
		trailerHeaders.Set("X-Content-SHA256", checksum)
		trailerHeaders.Set("X-Content-Length", contentLength)

		err = w.WriteTrailers(trailerHeaders)
		if err != nil {
			log.Println(err)
			return internalError
		}

		return nil
	}

	body := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	err := w.WriteStatusLine(response.StatusCodeOK)
	if err != nil {
		return internalError
	}

	h := headers.GetDefaultHeaders(len(body))
	h.Overwrite("Content-Type", "text/html")
	err = w.WriteHeaders(h)
	if err != nil {
		return internalError
	}

	_, err = w.WriteBody([]byte(body))
	if err != nil {
		return internalError
	}
	return nil
}
