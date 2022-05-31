package main

import (
	"errors"
	"fmt"
	"gnet-server/gnet"
	"net"
	"strings"
	"os"
	"syscall"
	"os/signal"
)

type ghandler struct {
	name string
}

func (h *ghandler) OnConnect(c gnet.Connector, reconnect bool) {
	fmt.Println("OnConnect: ", c.RemoteAddr())
}	


func (h *ghandler) OnDisconnect(c gnet.Connector) {
	fmt.Println("OnDisconnect: ", c.RemoteAddr())
}


func (h *ghandler) OnPacket(c gnet.Connector, packet []byte) error {

	if packet == nil {
		return nil
	}

	fmt.Println("OnPacket: ", c.RemoteAddr(), string(packet))
	c.SendPacket(packet)
	return nil
}

type EchoAdaption struct {
	RawConn net.Conn
}

func NewEchoAdaption(conn net.Conn) *EchoAdaption {
	ea := &EchoAdaption{
		RawConn : conn,
	}

	return ea
}

func (e *EchoAdaption) ReadMessage(sessionID int64) ([]byte, error) {
	buffer := make([]byte, 1024)
	n, err := e.RawConn.Read(buffer)
	if err != nil {
		return nil, err
	}

	if n <= 0 {
		return nil, err
	}

	return buffer[:n], nil
}

func (e *EchoAdaption) WriteMessage(data []byte) error {
	_, err := e.RawConn.Write(data)
	return err
}

func (e *EchoAdaption) Close() error {
	err := e.RawConn.Close()
	return err
}

func (e *EchoAdaption) Ping(msg []byte) error {
	return nil
}

func (e *EchoAdaption) LocalAddr() net.Addr {
	return e.RawConn.LocalAddr()
}

func (e *EchoAdaption) RemoteAddr() net.Addr {
	return e.RawConn.RemoteAddr()
}



type tcpserver struct {
	hub *gnet.Hub
	quitC chan struct{}
	handler gnet.NetHandler
	listen *net.TCPListener
}

func NewTCPServer(addr string, hub *gnet.Hub, handler gnet.NetHandler) (*tcpserver, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	listen, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}

	if hub == nil {
		hub = gnet.NewHub()
	}

	ts := &tcpserver{
		hub: hub,
		quitC: make(chan struct{}),
		handler: handler,
		listen: listen,
	}

	return ts, nil
}

func (s *tcpserver) Serve(recvChanSize, writeChanSize int) error {
	defer s.Stop()

	for {
		conn, err := s.listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}

			return err
		}

		op := NewEchoAdaption(conn)
		sp := gnet.NewServerPeer(op, s.hub, recvChanSize, writeChanSize)

		sp.Serve(s.handler)
	}

	return nil
}

func (s *tcpserver) Stop() error {
	select {
	case <-s.quitC:
		return errors.New("TcpServer already closed")
	default:
		close(s.quitC)
		s.hub.Close()
		s.listen.Close()

		return nil
	}
}




func main() {

	h := &ghandler{
		name: "chengmiao",
	}

	s, err := NewTCPServer("localhost:5000", nil, h)
	if err != nil {
		fmt.Println("NewTCPServer error: ", err)
		return
	}

	go s.Serve(1024, 1024)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}