package gnet

import (
	"sync"
)

// Hub Connection
type Hub struct {
	rwlock       *sync.RWMutex
	Connections  map[int64]Connector
	clientCloseC chan int64 // 关闭的客户端队列
	exitC        chan struct{}
}

// _Add 添加
func (m *Hub) _Add(id int64, c Connector) {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()
	m.Connections[id] = c
}

// _Del 删除
func (m *Hub) _Del(id int64) {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()
	delete(m.Connections, id)
}

// At get a Connection
func (m *Hub) Find(id int64) Connector {
	m.rwlock.RLock()
	defer m.rwlock.RUnlock()
	if conn, ok := m.Connections[id]; ok {
		return conn
	}
	return nil
}

// Traverse all Connection
func (m *Hub) Traverse(f func(Connector)) {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	for _, c := range m.Connections {
		f(c)
	}
}

// Multicast msg of sessionIDs
func (m *Hub) Multicast(group []int64, pkt []byte) (n int) {
	m.rwlock.RLock()
	defer m.rwlock.RUnlock()
	for _, id := range group {
		if conn, ok := m.Connections[id]; ok {
			conn.SendPacket(pkt)
			n++
		}
	}
	return
}

// Boardcast msg of sessionIDs
func (m *Hub) Boardcast(pkt []byte) (n int) {
	m.rwlock.RLock()
	defer m.rwlock.RUnlock()
	for _, conn := range m.Connections {
		conn.SendPacket(pkt)
		n++
	}
	return
}

// CloseClient close some client
func (m *Hub) CloseClient(id int64) {
	m.clientCloseC <- id
}

// Close self
func (m *Hub) Close() {
	select {
	case <-m.exitC:
		return
	default:
		close(m.exitC)

		m.rwlock.Lock()
		defer m.rwlock.Unlock()

		for _, v := range m.Connections {
			v.Close()
		}
	}
}

func (m *Hub) serve() {
	go func() {
		for {
			select {
			case <-m.exitC:
				return
			case id := <-m.clientCloseC:
				m._Del(id)
			}
		}
	}()
}

// NewHub create
func NewHub() *Hub {
	hub := &Hub{
		rwlock:       new(sync.RWMutex),
		Connections:  make(map[int64]Connector),
		clientCloseC: make(chan int64, 100),
		exitC:        make(chan struct{}),
	}
	hub.serve()
	return hub
}
