package tcp

import (
	"errors"
	"net"
	"gnet"
	"time"
)

// TCPClient use BinaryHeader, implement reconnect
type TCPClient struct {
	addr      string
	quitC     chan struct{}
	handler   gnet.NetHandler
	peer      *gnet.ClientPeer
	keepalive bool
}

func (c *TCPClient) serve(encryword byte) {
	reconn := false
LOOP:
	ntime := time.Now().Second()

	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		if c.keepalive {
			n := time.Now().Second() - ntime
			if n < 9 {
				time.Sleep(10 * time.Second)
			}
			reconn = true
			goto LOOP
		}
		return
	}
	op := NewAdapter(conn, encryword)

	c.quitC = make(chan struct{})

	c.peer = gnet.NewClientPeer(op)
	c.peer.Serve(c.handler, reconn)

	// reconnect
	c.peer = nil
	select {
	case <-c.quitC:
		return
	default:
		if c.keepalive {
			n := time.Now().Second() - ntime
			if n < 9 {
				time.Sleep(10 * time.Second)
			}
			reconn = true
			goto LOOP
		}
	}
}

// Stop manual close
func (c *TCPClient) Stop() error {
	c.keepalive = false
	return c.Close()
}

// Close manual close
func (c *TCPClient) Close() error {
	if c.peer != nil {
		select {
		case <-c.quitC:
			return errors.New("TCP Client has stopped")
		default:
			close(c.quitC)
			c.keepalive = false
			c.peer.Close()
			return nil
		}
	} else {
		return errors.New("TCP Client not started")
	}
}

// NewClient create a tcp Client
func NewClient(addr string, handler gnet.NetHandler, encryword byte, keepalive bool) (*TCPClient, error) {
	cli := &TCPClient{
		addr:      addr,
		handler:   handler,
		keepalive: keepalive}

	go cli.serve(encryword)
	return cli, nil
}
