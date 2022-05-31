package gnet

import (
	"net"
	"time"
)

var (
	// HeartbeatTime client Heartbeat tick
	HeartbeatTime = time.Second * 60

	CloseNormalClosure = 1000
)

// Adapter Interface
type Adapter interface {
	ReadMessage(sessionID int64) (data []byte, err error)
	WriteMessage(data []byte) error
	Ping(msg []byte) error
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}
