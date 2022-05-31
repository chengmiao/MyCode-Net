package ws

import (
	"errors"
	"gnet"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Adaption websocket net package Adapter,for github.com/gorilla/websocket
type Adaption struct {
	RawConn *websocket.Conn
}

// ReadMessage is a helper method for getting a reader using rawConn
func (m *Adaption) ReadMessage() ([]byte, error) {
	err := m.RawConn.SetReadDeadline(time.Now().Add(gnet.HeartbeatTime))
	if err != nil {
		return nil, err
	}
	mt, p, err := m.RawConn.ReadMessage() //
	if err == nil {
		if mt != websocket.BinaryMessage {
			return nil, errors.New("unsupported data format") // text protocol is not supported
		}
	}
	return p, err
}

// WriteMessage is a helper method for getting a writer using rawConn
func (m *Adaption) WriteMessage(data []byte) error {
	err := m.RawConn.SetWriteDeadline(time.Now().Add(gnet.HeartbeatTime))
	if err != nil {
		return err
	}
	return m.RawConn.WriteMessage(websocket.BinaryMessage, data)
}

// Ping contorl message
func (m *Adaption) Ping(msg []byte) error {
	return m.RawConn.WriteControl(websocket.PingMessage, msg, time.Now().Add(gnet.HeartbeatTime))
}

// LocalAddr returns the local network address
func (m *Adaption) LocalAddr() net.Addr {
	return m.RawConn.LocalAddr()
}

// RemoteAddr returns the remote network address
func (m *Adaption) RemoteAddr() net.Addr {
	return m.RawConn.RemoteAddr()
}

// Close the connect
func (m *Adaption) Close() error {
	p := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	if err := m.RawConn.WriteControl(websocket.CloseMessage, p, time.Now().Add(gnet.HeartbeatTime)); err != nil {
		return err
	}
	return m.RawConn.Close()
}

func newAdaption(conn *websocket.Conn) *Adaption {
	m := &Adaption{RawConn: conn}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(gnet.HeartbeatTime))
	})

	conn.SetPingHandler(func(message string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(gnet.HeartbeatTime))
		if err == nil {
			return conn.SetReadDeadline(time.Now().Add(gnet.HeartbeatTime))
		}
		return err
	})
	return m
}
