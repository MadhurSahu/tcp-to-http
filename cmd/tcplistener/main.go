package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		ch := getLinesChannel(conn)

		for line := range ch {
			fmt.Println(line)
		}

		fmt.Println("Connection to", conn.RemoteAddr(), "closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	buffer := ""

	go func() {
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)

			if err != nil {
				if buffer != "" {
					ch <- buffer
				}

				if err == io.EOF {
					close(ch)
					return
				}

				log.Fatal(err)
			}

			buffer += string(data[:n])
			lines := strings.Split(buffer, "\n")

			for i := 0; i < len(lines)-1; i++ {
				ch <- lines[i]
			}

			buffer = lines[len(lines)-1]
		}
	}()

	return ch
}
