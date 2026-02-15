package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {

	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Println("unable to resolve the network name")
		return
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("connection failed")
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s ", ">")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error reading the line")
		}
		conn.Write([]byte(line))
	}

}
