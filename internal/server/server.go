package server

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/phungducminh/httpfromtcp/internal/request"
	"github.com/phungducminh/httpfromtcp/internal/response"
)

type Server struct {
	h        Handler
	listener net.Listener

	mu          sync.RWMutex
	connections map[net.Conn]struct{}

	closed atomic.Bool
}

// Serve create a net.Listener and returns a new Server instance. It also start
// listening for requests in a goroutine
func Serve(port int, h Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener:    listener,
		connections: map[net.Conn]struct{}{},
		closed:      atomic.Bool{},
		h:           h,
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
		slog.Error("failed to write to connection", slog.Any("err", err))
	}

	b := &bytes.Buffer{}
	hErr := s.h(b, req)
	if hErr != nil {
		HandleError(conn, hErr)
		return
	}

	response.WriteStatusLine(conn, response.OK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(b.Len()))
	err = response.WriteBody(conn, b.Bytes())
	if err != nil {
		if !s.closed.Load() {
			slog.Error("failed to write to connection", slog.Any("err", err))
		} else {
			slog.Info("failed to write to connection as server has already been closed")
		}
	}
}

func HandleError(w io.Writer, hErr *HandlerError) {
	message := hErr.Message
	response.WriteStatusLine(w, hErr.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(message)))
	err := response.WriteBody(w, []byte(message))
	if err != nil {
		slog.Error("failed to write to connection", slog.Any("err", err))
	}
}
