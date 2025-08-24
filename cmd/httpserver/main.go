package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/phungducminh/httpfromtcp/internal/request"
	"github.com/phungducminh/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	var h server.Handler = func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return server.NewHandlerError(400, "Your problem is not my problem\n")
		case "/myproblem":
			return server.NewHandlerError(500, "Woopsie, my bad\n")
		}
		w.Write([]byte("All good, frfr\n"))
		return nil
	}
	server, err := server.Serve(port, h)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
