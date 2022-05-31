package ws

import (
	"context"
	"errors"
	"gnet"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"
)

// WSServer websocket 的第三方版本
type WSServer struct {
	*gnet.Hub
	quitC   chan struct{}
	handler gnet.NetHandler
	server  *http.Server
}

// Serve : continuous service
func (ws *WSServer) Serve() {
	ws.server.ListenAndServe()
}

// Stop the WebSocket Server
func (ws *WSServer) Stop() error {
	select {
	case <-ws.quitC:
		return errors.New("WebSocket Service has stopped")
	default:
		close(ws.quitC)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return ws.server.Shutdown(ctx)
	}
}

// NewServer Create a Server
func NewServer(addr, pattern string, hub *gnet.Hub, hd gnet.NetHandler) (*WSServer, error) {
	if hub == nil {
		hub = gnet.NewHub() // use private Hub
	}

	var upgrader = websocket.Upgrader{}

	mux := &http.ServeMux{}

	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		op := newAdaption(c)
		peer := gnet.NewServerPeer(op, hub)
		peer.Serve(hd)
		peer.WaitClose()
	})

	httpserver := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	svr := &WSServer{
		Hub:     hub,
		handler: hd,
		server:  httpserver,
		quitC:   make(chan struct{}),
	}
	log.Printf("websocket server(%v%v) start...", addr, pattern)
	return svr, nil
}
