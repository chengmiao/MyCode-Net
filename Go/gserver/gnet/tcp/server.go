package tcp

import (
	"errors"
	"gnet"
	"net"
	"strings"

	log "github.com/rs/zerolog/log"
)

// TCPServer use BinaryHeader, that stutters ... :)
type TCPServer struct {
	*gnet.Hub
	encryword byte
	quitC     chan struct{}
	handler   gnet.NetHandler
	listen    *net.TCPListener
}

// Serve : continuous service
func (s *TCPServer) Serve(recvChanSize, writeChanSize int) error {
	defer s.Stop()
	for {
		conn, err := s.listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				break
			}
			log.Fatal().AnErr("tcp Accept Error:", err)
			return err
		}

		op := NewAdapter(conn, s.encryword)

		peer := gnet.NewServerPeer(op, s.Hub, recvChanSize, writeChanSize)
		peer.Serve(s.handler)
	}
	return nil
}

// Stop the Server
func (s *TCPServer) Stop() error {
	select {
	case <-s.quitC:
		return errors.New("TCP Service has stopped")
	default:
		close(s.quitC)
		s.Hub.Close()
		s.listen.Close()

		return nil
	}
}

func (s *TCPServer) GetHandler() gnet.NetHandler {
	return s.handler
}

// NewServer create a tcp server
func NewServer(addr string, hub *gnet.Hub, handler gnet.NetHandler, encryword byte) (*TCPServer, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	listen, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}
	log.Trace().Str("Addr", addr).Msg("Tcp server start")
	if hub == nil {
		hub = gnet.NewHub() // use private Hub
	}

	srv := &TCPServer{
		Hub:       hub,
		handler:   handler,
		listen:    listen,
		encryword: encryword,
		quitC:     make(chan struct{}),
	}

	return srv, err
}
