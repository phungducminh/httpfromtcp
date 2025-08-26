package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/phungducminh/httpfromtcp/internal/headers"
	"github.com/phungducminh/httpfromtcp/internal/request"
	"github.com/phungducminh/httpfromtcp/internal/response"
	"github.com/phungducminh/httpfromtcp/internal/server"
)

const port = 42069

func respond200() string {
	return `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
}

func respond400() string {
	return `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
}

func respond500() string {
	return `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
}

func main() {
	var h server.Handler = func(w *response.Writer, req *request.Request) {
		h := headers.NewHeaders()
		h.Replace("Content-Type", "text/html")
		body := respond200()
		status := 200
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body = respond400()
			status = 400
		case "/myproblem":
			body = respond500()
			status = 500
		}
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteStatusLine(response.StatusCode(status))
		w.WriteHeaders(h)
		w.WriteBody([]byte(body))
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
