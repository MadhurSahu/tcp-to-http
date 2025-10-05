package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ch := getLinesChannel(file)

	for line := range ch {
		fmt.Printf("read: %s\n", line)
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
