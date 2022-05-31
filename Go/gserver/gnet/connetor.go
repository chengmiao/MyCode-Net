package gnet

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// NetHandler 外部网络事件处理器
type NetHandler interface {
	OnConnect(Connector, bool)
	OnDisconnect(Connector)
	OnPacket(Connector, []byte) error // 接收到完整封包
}

// Connector Peer interface
type Connector interface {
	SendPacket([]byte) error
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SessionID() int64
	SetUserData(data interface{})
	GetUserData() interface{}
}

// ServerConnector connect of server
type ServerConnector interface {
	SessionID() int64
	SendMessage(int64, []byte) error
}

// ServerPeer Connector of server side
type ServerPeer struct {
	innerConn Adapter
	readC     chan []byte
	writeC    chan []byte
	exitC     chan struct{}
	sessionID int64
	hub       *Hub
	wg        sync.WaitGroup
	once      sync.Once
	userData  interface{}
}

// SetUserData impl
func (sp *ServerPeer) SetUserData(data interface{}) {
	sp.userData = data
}

// GetUserData impl
func (sp *ServerPeer) GetUserData() interface{} {
	return sp.userData
}

// notify read/write goroutine closed
func (sp *ServerPeer) notifyClose(err error) {
	sp.once.Do(func() {
		log.Info().Int64("SessionID", sp.sessionID).Err(err).Msg("Session close.")
		close(sp.exitC)
		sp.hub.CloseClient(sp.sessionID)
	})
}

// LocalAddr impl
func (sp *ServerPeer) LocalAddr() net.Addr {
	return sp.innerConn.LocalAddr()
}

type test struct {
	serverID   uint32
	serverType uint32
}

// 让 NetHandler 在同一个goroute
func (sp *ServerPeer) readLoop(ev NetHandler) {
	defer func() {
		ev.OnDisconnect(sp)
		sp.wg.Done()
	}()

	ev.OnConnect(sp, false)
	for {
		select {
		case <-sp.exitC:
			return
		default:
			p, err := sp.innerConn.ReadMessage(sp.sessionID)
			if err != nil {
				sp.notifyClose(err)
				sp.innerConn.Close()
				return
			}

			err = ev.OnPacket(sp, p)
			if err != nil {
				sp.notifyClose(err)
				sp.innerConn.Close()
				return
			}
		}
	}
}

func (sp *ServerPeer) sendLoop() {
	defer sp.wg.Done()
	for {
		select {
		case <-sp.exitC:
			return
		case pkt := <-sp.writeC:
			if err := sp.innerConn.WriteMessage(pkt); err != nil {
				sp.notifyClose(err) // write failed
				return
			}
		}
	}
}

// Serve 启动服务
func (sp *ServerPeer) Serve(ev NetHandler) {
	select {
	case <-sp.exitC:
		return
	default:
		sp.wg.Add(2)
		go sp.readLoop(ev)
		go sp.sendLoop()
		// serverPeer is async, don`t sp.wg.Wait()
	}
}

// IsClosed 是否已经关闭
func (sp *ServerPeer) IsClosed() bool {
	select {
	case <-sp.exitC:
		return true
	default:
		return false
	}
}

// SessionID 会话编号 Connector interface
func (sp *ServerPeer) SessionID() int64 {
	return sp.sessionID
}

// SendPacket 发送 Connector interface
func (sp *ServerPeer) SendPacket(pkt []byte) error {
	select {
	case <-sp.exitC:
		return errors.New("Cannot use a closed ServerPeer")
	case sp.writeC <- pkt:
		return nil
		// case <-time.After(time.Millisecond * 200):
		// 	return errors.New("The async send buffer is full")
	}
}

// Close 主动关闭 Connector interface
func (sp *ServerPeer) Close() error {
	err := sp.innerConn.Close() // 原生的连接关闭,等待sendLoop readloop退出
	sp.wg.Wait()
	return err
}

// RemoteAddr  Connector interface
func (sp *ServerPeer) RemoteAddr() net.Addr {
	return sp.innerConn.RemoteAddr()
}

// WaitClose 等待关闭
func (sp *ServerPeer) WaitClose() {
	sp.wg.Wait()
}

// NewServerPeer create a ServerPeer
func NewServerPeer(op Adapter, h *Hub, recvChanSize, writeChanSize int) *ServerPeer {
	id := createSessionID()
	c := &ServerPeer{
		innerConn: op, // hide method of the Conn
		readC:     make(chan []byte, recvChanSize),
		writeC:    make(chan []byte, writeChanSize),
		exitC:     make(chan struct{}),
		hub:       h,
		sessionID: id,
	}
	if h != nil {
		h._Add(id, c)
	}
	return c
}

// ClientPeer Client Socket
type ClientPeer struct {
	innerConn Adapter
	readC     chan []byte
	writeC    chan []byte
	exitC     chan struct{}
	wg        sync.WaitGroup
	once      sync.Once
	sessionID int64
	userData  interface{}
}

// SetUserData impl
func (cp *ClientPeer) SetUserData(data interface{}) {
	cp.userData = data
}

// GetUserData impl
func (cp *ClientPeer) GetUserData() interface{} {
	return cp.userData
}

// SendPacket Connector interface
func (cp *ClientPeer) SendPacket(pkt []byte) error {
	select {
	case <-cp.exitC:
		return errors.New("Cannot use a closed ServerPeer")
	case cp.writeC <- pkt:
		return nil
		// case <-time.After(time.Millisecond * 200):
		// 	return errors.New("The async send buffer is full")
	}
}

// SessionID Connector interface
func (cp *ClientPeer) SessionID() int64 {
	return cp.sessionID
}

// Close Connector interface
func (cp *ClientPeer) Close() error {
	err := cp.innerConn.Close()
	cp.wg.Wait()
	return err
}

// LocalAddr Connector interface
func (cp *ClientPeer) LocalAddr() net.Addr {
	return cp.innerConn.LocalAddr()
}

// RemoteAddr Connector interface
func (cp *ClientPeer) RemoteAddr() net.Addr {
	return cp.innerConn.RemoteAddr()
}

// Serve 启动服务
func (cp *ClientPeer) Serve(ev NetHandler, reconnect bool) {
	select {
	case <-cp.exitC:
		return
	default:
		cp.wg.Add(2)
		go cp.readLoop(ev, reconnect)
		go cp.sendLoop()
		cp.wg.Wait() // sync
	}
}

// notify read/write goroutine closed
func (cp *ClientPeer) notifyClose(err error) {
	cp.once.Do(func() {
		close(cp.exitC)
	})
}

// 让 NetHandler 在同一个goroute
func (cp *ClientPeer) readLoop(ev NetHandler, isReconnect bool) {
	defer func() {
		ev.OnDisconnect(cp)
		cp.wg.Done()
	}()

	ev.OnConnect(cp, isReconnect)
	for {
		select {
		case <-cp.exitC:
			return
		default:
			p, err := cp.innerConn.ReadMessage(cp.sessionID)
			if err != nil {
				cp.notifyClose(err)
				cp.innerConn.Close()
				return
			}

			err = ev.OnPacket(cp, p)
			if err != nil {
				cp.notifyClose(err)
				cp.innerConn.Close()
				return
			}
		}
	}
}

func (cp *ClientPeer) sendLoop() {
	defer cp.wg.Done()
	ticker := time.NewTicker(HeartbeatTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := cp.innerConn.Ping([]byte{}) // heartbeat
			if err != nil {
				cp.notifyClose(err) // write failed
				return
			}
		case <-cp.exitC:
			return
		case pkt := <-cp.writeC:
			if err := cp.innerConn.WriteMessage(pkt); err != nil {
				cp.notifyClose(err) // write failed
				return
			}
		}
	}
}

// NewClientPeer create a ClientPeer
func NewClientPeer(op Adapter) *ClientPeer {
	id := createSessionID()
	c := &ClientPeer{
		innerConn: op,
		readC:     make(chan []byte, 100),
		writeC:    make(chan []byte, 100),
		exitC:     make(chan struct{}),
		sessionID: id,
	}
	return c
}
