package ws

import (
	"errors"
	"gnet"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"
)

// WSClient use BinaryHeader, implement reconnect
type WSClient struct {
	addr      string
	quitC     chan struct{}
	handler   gnet.NetHandler
	peer      *gnet.ClientPeer
	reconnect bool
}

func (c *WSClient) serve(urlstr string) {
	reconn := false
LOOP:
	ntime := time.Now().Second()
	conn, _, err := websocket.DefaultDialer.Dial(urlstr, nil)
	if err != nil {
		return
	}

	c.quitC = make(chan struct{})
	op := newAdaption(conn)
	c.peer = gnet.NewClientPeer(op)
	c.peer.Serve(c.handler, reconn)
	c.peer = nil
	select {
	case <-c.quitC:
		return
	default:
		if c.reconnect {
			n := time.Now().Second() - ntime
			if n < 9 {
				time.Sleep(10 * time.Second)
			}
			reconn = true
			goto LOOP
		}
	}
}

// Stop the Client
func (c *WSClient) Stop() error {
	c.reconnect = false
	return c.Close()
}

// Close manual close
func (c *WSClient) Close() error {
	if c.peer != nil {
		select {
		case <-c.quitC:
			return errors.New("TCP Client has stopped")
		default:
			c.reconnect = false
			close(c.quitC)
			c.peer.Close()
			return nil
		}
	} else {
		return errors.New("TCP Client not started")
	}
}

// NewClient create a tcp Client
func NewClient(addr, path string, handler gnet.NetHandler, reconnect bool) (*WSClient, error) {
	cli := &WSClient{
		addr:      addr,
		handler:   handler,
		reconnect: reconnect}
	u := url.URL{Scheme: "ws", Host: addr, Path: path}
	log.Printf("connecting to %s", u.String())
	go cli.serve(u.String())
	return cli, nil
}
