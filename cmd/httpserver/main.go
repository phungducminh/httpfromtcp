package main

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	logLvl := flag.String("log-level", "INFO", "log level")

	flag.Parse()

	switch *logLvl {
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		panic("log level must be either DEBUG, INFO, WARN, ERROR")
	}

	var h server.Handler = func(w *response.Writer, req *request.Request) {
		h := headers.NewHeaders()
		h.Replace("Content-Type", "text/html")
		body := []byte(respond200())
		status := 200
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/yourproblem") {
			body = []byte(respond400())
			status = 400
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/myproblem") {
			body = []byte(respond500())
			status = 500
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			handleHttpBinRequest(req, w)
			return
		} else if req.RequestLine.RequestTarget == "/video" && req.RequestLine.Method == "GET" {
			h.Replace("Content-Type", "video/mp4")
			p, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				w.WriteInternalServerError(err, h)
				return
			}
			body = p
		}
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteStatusLine(response.StatusCode(status))
		w.WriteHeaders(h)
		w.WriteBody(body)
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

func handleHttpBinRequest(req *request.Request, w *response.Writer){
	h := headers.NewHeaders()
	h.Replace("Content-Type", "text/html")
	suffix := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	res, err := http.Get("https://httpbin.org/" + suffix)
	if err != nil {
		w.WriteInternalServerError(err, h)
		return 
	}

	w.WriteStatusLine(response.OK)
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	h.Delete("Content-Length")
	w.WriteHeaders(h)

	hash256 := sha256.New()
	size := 0
	p := make([]byte, 32)
	for {
		eof := false
		n, err := res.Body.Read(p)
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = true
			} else {
				slog.Error("failed to read request body", slog.Any("error", err))
				// maybe return will close connection
				return
			}
		}
		if n == 0 {
			break
		}

		slog.Debug("write chunk body", slog.String("data", string(p[:n])))
		n, _ = hash256.Write(p[:n])
		size += n
		w.WriteChunkedBody(p[:n])
		if eof {
			break
		}
	}
	slog.Debug("write chunk body done")
	w.Write([]byte("0\r\n"))
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", string(hash256.Sum(nil))))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", size))
	w.WriteTrailers(trailers)
}
