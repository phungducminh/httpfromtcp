package server

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/phungducminh/httpfromtcp/internal/request"
)

type Server struct {
	listener net.Listener

	mu          sync.RWMutex
	connections map[net.Conn]struct{}

	closed atomic.Bool
}

// Serve create a net.Listener and returns a new Server instance. It also start
// listening for requests in a goroutine
func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener:    listener,
		connections: map[net.Conn]struct{}{},
		closed:      atomic.Bool{},
	}
	go s.listen()
	return s, nil
}

// Close close the listeners and the server
func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		slog.Error("failed to close listener", slog.Any("err", err))
	}

	s.closed.Store(true)

	return nil
}

// listen uses a loop to accept new connections as they come in and handle each
// one in a new goroutine
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", slog.Any("err", err))
		}
		s.mu.Lock()
		s.connections[conn] = struct{}{}
		s.mu.Unlock()
		go s.handle(conn)
	}
}

// handle handles a single connection and then closes the connection
func (s *Server) handle(conn net.Conn) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
	}()

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
	// TODO: @minh reclaim non-used allocated memory
	b := bytes.Buffer{}
	b.WriteString("HTTP/1.1 200 OK\r\n")
	b.WriteString("Content-Type: text/plain\r\n")
	b.WriteString("Content-Length: 12\r\n")
	b.WriteString("\r\n")
	b.WriteString("Hello World!")
	slog.Info("send response", slog.String("message", b.String()))

	_, err = conn.Write(b.Bytes())
	if err != nil {
		if !s.closed.Load() {
			slog.Error("failed to write to connection", slog.Any("err", err))
		} else {
			slog.Info("failed to write to connection as server has already been closed")
		}
	}
}
