package main

import (
	"fmt"
	"net"
	"strings"
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
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("%s\n", line)
		}
	}

}

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		buf := make([]byte, 8)
		var currentLine string = ""

		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			split := strings.Split(string(buf[:n]), "\n")

			if len(split) > 1 {
				lines <- currentLine + split[0]
				currentLine = split[1]
			} else {
				currentLine += split[0]
			}
		}
		if currentLine != "" {
			lines <- currentLine
		}
	}()
	return lines
}
