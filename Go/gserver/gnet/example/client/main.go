package main

import (
	"log"
	"runtime/pprof"
	"gnet/tcp"
	"gnet/ws"
	"time"

	_ "net/http/pprof"
	"os"
	"os/signal"
	"gnet"
	"sync/atomic"
	"syscall"
)

const defaultEncryptionWord = 5 // 0~31

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	clihandle := &myHandle{"client", 0}
	// ================== tcp ======================= //
	for i := 0; i < 3000; i++ {
		_, err := tcp.NewClient("10.246.34.79:8081", clihandle, defaultEncryptionWord, false)
		//		defer s11.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}
	// ==================== websocket ===================== //
	for i := 0; i < 3000; i++ {
		_, err := ws.NewClient("10.246.34.79:8080", "/", clihandle, false)
		if err != nil {
			log.Fatalln(err)
		}
	}
	// -------- block signal --------------- //
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// -------- prof ----------------------- //
	f, err := os.OpenFile("./cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	timeout := time.After(time.Second * 30)
	select {
	case <-sig:
		break
	case <-timeout:
		log.Println(clihandle.count / 30.0)
		break
	}

	pprof.StopCPUProfile()
	f.Close()

}

type myHandle struct {
	name  string
	count int32
}

func (m *myHandle) OnConnect(c gnet.Connector, reconnect bool) {
	// log.Println(m.name, "OnConnect")
	for i := 0; i < 10; i++ {
		c.SendPacket([]byte("ok === ok, " + m.name))
	}
}
func (m *myHandle) OnDisconnect(c gnet.Connector) {
	// log.Println(m.name, "OnDisconnect")
}

func (m *myHandle) OnPacket(c gnet.Connector, buf []byte) error {
	// msg := string(buf)
	// log.Println(m.name, "OnPacket", msg)
	atomic.AddInt32(&m.count, 1)
	c.SendPacket(buf)
	return nil
}
