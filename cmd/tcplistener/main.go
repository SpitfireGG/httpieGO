package main

import (
	"fmt"
	"net"
	"os"

	"spitfiregg.httpFromScratch.httpieee/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("TCP server listening on :42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		fmt.Println("Connection accepted")

		r, err := request.RequestFromReader(conn)
		fmt.Println("Request Line")
		fmt.Printf("- Method: GET :: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: / :: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: 1.1 :: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Connection closed")
	}
}
