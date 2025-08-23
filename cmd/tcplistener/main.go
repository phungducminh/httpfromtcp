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

	fmt.Printf("Request line:\n")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

	fmt.Printf("Headers:\n")
	req.Headers.ForEach(func(key, value string) {
		fmt.Printf("- %s: %s\n", key, value)
	})

	fmt.Printf("Body:\n")
	fmt.Printf("%s\n", string(req.Body))
}
