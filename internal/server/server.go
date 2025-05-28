package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	Port     int
	Listener net.Listener
	Closed   atomic.Bool
}

func Serve(port int) (*Server, error) {

	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}

	server := &Server{Port: port, Listener: ln}
	server.Closed.Store(false)

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) listen() {
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
		go s.handle(conn)
	}
}

const resp = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\r\n"

func (s *Server) handle(conn net.Conn) {
	// req, err := request.RequestFromReader(conn)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Request line:")
	// fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	// fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	// fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	// fmt.Println("Headers:")
	// for key, value := range req.Headers {
	// 	fmt.Printf("- %s: %s\n", key, value)
	// }
	// fmt.Println("Body:")
	// if len(req.Body) > 0 {
	// 	fmt.Println(string(req.Body))
	// } else {
	// 	fmt.Println("(no body)")
	// }

	conn.Write([]byte(resp))
	fmt.Println("Response sent")

	conn.Close()
}
