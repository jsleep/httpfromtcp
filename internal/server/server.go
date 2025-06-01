package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/jsleep/httpfromtcp/internal/request"
	"github.com/jsleep/httpfromtcp/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	Closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {

	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}

	server := &Server{Port: port, Listener: ln}
	server.Closed.Store(false)

	go server.listen(handler)

	return server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) listen(handler Handler) {
	for {
		if s.Closed.Load() {
			fmt.Println("Server is closed, stopping listener")
			return
		}
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("accepted connection")
		go s.handle(conn, handler)
	}
}

const respBody = "Hello World!\r\n"

func (s *Server) handle(conn net.Conn, handler Handler) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
	}

	req.Print()

	w := response.Writer{Writer: conn}

	handlerError := handler(w, req)

	if handlerError != nil {
		fmt.Printf("Error handling request: %v\n", handlerError)
		return
	}

	conn.Close()
}

type Handler func(w response.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

// func (he *HandlerError) Error(w io.Writer) {
// 	response.WriteStatusLine(w, he.Code)
// 	response.WriteHeaders(w, response.GetDefaultHeaders(len(he.Message)))
// 	response.WriteBody(w, []byte(he.Message))
// }
