package main

import (
	"log"
	"os"
	"os/signal"
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
	internalError := &server.HandlerError{
		StatusCode: response.StatusCodeInternalServerError,
	}

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.StatusCodeBadRequest,
		}
	case "/myproblem":
		return internalError
	default:
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

		defaultHeaders := headers.GetDefaultHeaders(len(body))
		defaultHeaders.Overwrite("Content-Type", "text/html")
		err = w.WriteHeaders(defaultHeaders)
		if err != nil {
			return internalError
		}

		_, err = w.WriteBody([]byte(body))
		if err != nil {
			return internalError
		}
		return nil
	}
}
