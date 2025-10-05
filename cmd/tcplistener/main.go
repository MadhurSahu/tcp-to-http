package main

import (
	"fmt"
	"log"
	"net"

	"github.com/MadhurSahu/tcp-to-http/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	fmt.Println("Listening on port 42069")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connection accepted from", conn.RemoteAddr())

		req, err := request.FromReader(conn)
		if err != nil {
			log.Println(err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", req.RequestLine.Method)
		fmt.Println("- Target:", req.RequestLine.RequestTarget)
		fmt.Println("- Version:", req.RequestLine.HttpVersion)

		fmt.Println("Connection to", conn.RemoteAddr(), "closed")
	}
}
