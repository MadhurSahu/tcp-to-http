package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/MadhurSahu/tcp-to-http/internal/request"
	"github.com/MadhurSahu/tcp-to-http/internal/response"
)

type Server struct {
	closed   atomic.Bool
	handler  Handler
	listener net.Listener
}

type HandlerError struct {
	StatusCode response.StatusCode
}

type Handler func(w *response.Writer, req *request.Request) *HandlerError

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		s.listener.Close()
	}
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		time.Sleep(50 * time.Millisecond)
		conn.Close()
	}()
	res := response.NewWriter(conn)

	req, err := request.FromReader(conn)
	if err != nil {
		err := res.WriteError(response.StatusCodeBadRequest)
		if err != nil {
			log.Println(err)
		}
		return
	}

	hErr := s.handler(res, req)
	if hErr != nil {
		err := res.WriteError(hErr.StatusCode)
		if err != nil {
			log.Println(err)
		}
		return
	}
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %f", err)
			continue
		}
		go s.handle(conn)
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	go server.listen()

	return server, nil
}
