package main

import (
	"io"
	"log"
	"net"
	"fmt"
)

func handler(c *net.UDPConn){
	defer func() {
		fmt.Println("Udp Gone")
		c.Close()
	}()

	for {
		buf := make([]byte, 1024)
		n, clientaddr, _ := c.ReadFromUDP(buf)
		fmt.Println(string(buf[:n]))

		c.WriteToUDP(buf[:n], clientaddr)
	}
}

func main() {
	addr := "localhost:5000"
	TcpListen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	defer TcpListen.Close()

	udpaddr, _ := net.ResolveUDPAddr("udp", "localhost:5000")
	UdpListen, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		log.Fatal(err)
	}

	go handler(UdpListen)

	for {	
		conn, err := TcpListen.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		fmt.Println("Accept Loop")

		go func(conn net.Conn) {
			fmt.Println("connect success")
			io.Copy(conn, conn)
		}(conn)

		//go handler(conn)
	}
}