package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/jsleep/httpfromtcp/internal/headers"
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

func proxyHandler(w response.Writer, req *request.Request) {
	base_url := "http://httpbin.org"
	x := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

	full_url := base_url + "/" + x
	log.Printf("Proxying request to %s", full_url)

	proxy_headers := response.GetDefaultHeaders(0)
	delete(proxy_headers, "Content-Length") // Remove Content-Length to avoid conflicts
	proxy_headers["Transfer-Encoding"] = "chunked"
	proxy_headers["Trailer"] = "X-Content-Sha256, X-Content-Length"

	resp, err := http.Get(full_url)

	if err != nil {
		log.Printf("Error fetching %s: %v", full_url, err)
		w.WriteStatusLine(500)
		w.WriteHeaders(proxy_headers)
		w.WriteBody([]byte("Error: " + err.Error()))
		return
	}

	if resp.StatusCode != 200 {
		log.Printf("Received non-200 response: %d", resp.StatusCode)
		w.WriteStatusLine(500)
		w.WriteHeaders(proxy_headers)
		w.WriteBody([]byte("Error: " + resp.Status))
		return
	}

	log.Printf("Received response from %s: %d", full_url, resp.StatusCode)
	w.WriteStatusLine(200)
	w.WriteHeaders(proxy_headers)

	bufferSize := 1024 // 1 KB buffer size
	buffer := make([]byte, bufferSize)
	buffering := true
	bytesRead := 0
	bytesWritten := 0

	for buffering {
		n, _ := resp.Body.Read(buffer[bytesRead:])

		if n == 0 {
			buffering = false
			break
		}

		bytesRead += n
		w.WriteChunkedBody(buffer[bytesWritten:bytesRead])

		bytesWritten += n

		if bytesRead == len(buffer) {
			// If the buffer is full, double its size
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

	}
	resp.Body.Close()
	w.WriteChunkedBodyDone()
	trailers := headers.NewHeaders()

	hashBytes := sha256.Sum256(buffer[:bytesWritten])

	trailers["X-Content-Sha256"] = hex.EncodeToString(hashBytes[:])
	trailers["X-Content-Length"] = strconv.Itoa(bytesWritten)
	w.WriteTrailers(trailers)
}

func videoHandler(w response.Writer, req *request.Request) {
	body, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Printf("Error reading video file: %v", err)
		w.WriteStatusLine(500)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		w.WriteBody([]byte("Error: " + err.Error()))
		return
	}

	headers := response.GetDefaultHeaders(len(body))
	headers["Content-Type"] = "video/mp4"

	w.WriteStatusLine(200)
	w.WriteHeaders(headers)
	w.WriteBody(body)
}

var myHandler = server.Handler(func(w response.Writer, req *request.Request) *server.HandlerError {

	var body string
	var code response.StatusCode
	if req.RequestLine.RequestTarget == "/video" {
		videoHandler(w, req)
		return nil
	} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		proxyHandler(w, req)
		return nil
	} else if req.RequestLine.RequestTarget == "/yourproblem" {
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
