package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("udp", ":42069")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			// if err.Error() == "EOF" {
			// 	break
			// }
			panic(err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println("Error sending data:", err)
			continue
		}
	}
}
