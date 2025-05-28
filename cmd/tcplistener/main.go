package main

import (
	"fmt"
	"net"

	"github.com/jsleep/httpfromtcp/internal/request"
)

func main() {

	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	// Read the file
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("accepted connection")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			panic(err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		if len(req.Body) > 0 {
			fmt.Println(string(req.Body))
		} else {
			fmt.Println("(no body)")
		}
	}

}
