package main

import (
	"fmt"
	"net"

	"github.com/phungducminh/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Printf("connection accepted\n")
		handleConnect(conn)
		fmt.Printf("connection closed\n")
	}
}

func handleConnect(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		panic(err)
	}

	fmt.Printf(`Request line:
- Method: %s
- Target: %s
- Version: %s`, req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)

	fmt.Printf("\nHeaders:\n")
	req.Headers.ForEach(func(key, value string) {
		fmt.Printf("- %s: %s\n", key, value)
	})
}
