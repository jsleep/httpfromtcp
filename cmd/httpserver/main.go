package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsleep/httpfromtcp/internal/request"
	"github.com/jsleep/httpfromtcp/internal/response"
	"github.com/jsleep/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, myHandler)
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

// HandlerError represents an HTTP error with a status code and message.
type HandlerError struct {
	StatusCode int
	Message    string
}

var myHandler = server.Handler(func(w response.Writer, req *request.Request) *server.HandlerError {

	var body string
	var code response.StatusCode

	if req.RequestLine.RequestTarget == "/yourproblem" {
		code = 400
		body =
			`<html>
				<head>
					<title>400 Bad Request</title>
				</head>
				<body>
					<h1>Bad Request</h1>
					<p>Your request honestly kinda sucked.</p>
				</body>
			</html>`
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		code = 500
		body = `<html>
					<head>
						<title>500 Internal Server Error</title>
					</head>
					<body>
						<h1>Internal Server Error</h1>
						<p>Okay, you know what? This one is on me.</p>
					</body>
				</html>`
	} else {
		code = 200
		body = `<html>
					<head>
						<title>200 OK</title>
					</head>
					<body>
						<h1>Success!</h1>
						<p>Your request was an absolute banger.</p>
					</body>
				</html>`
	}
	headers := response.GetDefaultHeaders(len(body))
	headers["Content-Type"] = "text/html"

	w.WriteStatusLine(code)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(body))
	return nil
})
