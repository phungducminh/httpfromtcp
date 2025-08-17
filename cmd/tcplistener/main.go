package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func getLinec(f io.ReadCloser) <-chan string {
	linec := make(chan string, 10)
	go func() {
		defer func() {
			f.Close()
			close(linec)
		}()

		var line string = ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			data = data[:n]
			if idx := bytes.IndexByte(data, '\n'); idx != -1 {
				line += string(data[:idx])
				linec <- line
				line = string(data[idx+1:])
			} else {
				line += string(data)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
		}

		if len(line) > 0 {
			linec <- line
		}
	}()
	return linec
}

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
		linec := getLinec(conn)
		for line := range linec {
			fmt.Printf("%s\n", line)
		}
		select {
		case _, ok := <-linec:
			if !ok {
				fmt.Printf("connection closed")
			}
		}
	}
}
