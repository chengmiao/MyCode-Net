package main

import (
	"fmt"
	"net"
	"time"
)

func main() {

	udpconn, _ := net.Dial("udp", "127.0.0.1:5000")
	tcpconn, _ := net.Dial("tcp", "127.0.0.1:5000")

	defer tcpconn.Close()
	defer udpconn.Close()

	for {

		tcpconn.Write([]byte("Hello TCP"))
		udpconn.Write([]byte("Hello UDP"))

		buffer := make([]byte, 1024)
		n, _ := tcpconn.Read(buffer)
		fmt.Println(string(buffer[:n]))

		n, _ = udpconn.Read(buffer)
		fmt.Println(string(buffer[:n]))


		time.Sleep(10 * time.Second)
	}
}