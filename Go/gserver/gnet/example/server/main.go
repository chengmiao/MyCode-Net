package main

import (
	"flag"
	"log"
	"runtime/pprof"
	"gnet/tcp"
	"gnet/ws"

	_ "net/http/pprof"
	"os"
	"os/signal"
	"gnet"
	"syscall"
)

const defaultEncryptionWord = 5

func main() {

	pServerid := flag.Int("id", 1, "server id")
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	gnet.SetServerID(*pServerid) //

	hub := gnet.NewHub() // public shared
	srvhandle := &myHandle{"server"}

	// =================== tcp server ====================== //

	s1, err := tcp.NewServer(":8081", hub, srvhandle, defaultEncryptionWord)
	if err != nil {
		log.Fatalln(err)
	}
	defer s1.Stop()
	go s1.Serve()

	// ====================== websocket server =================== //
	s2, err := ws.NewServer(":8080", "/", hub, srvhandle)
	if err != nil {
		log.Fatalln(err)
	}
	defer s2.Stop()
	go s2.Serve()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	f, err := os.OpenFile("./cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	<-sig

	pprof.StopCPUProfile()
	f.Close()

}

type myHandle struct {
	name string
}

func (m *myHandle) OnConnect(c gnet.Connector, reconnect bool) {
	// log.Println(m.name, "OnConnect")
	// c.SendPacket([]byte("ok === ok, " + m.name))
}
func (m *myHandle) OnDisconnect(c gnet.Connector) {
	// log.Println(m.name, "OnDisconnect")
}

func (m *myHandle) OnPacket(c gnet.Connector, buf []byte) error {
	c.SendPacket(buf)
	return nil
}
