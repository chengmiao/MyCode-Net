package main

import (
	"log"
	"runtime/pprof"
	"gnet/tcp"
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

	hub := gnet.NewHub()
	srvhandle := &myHandle{"server", 0}
	clihandle := &myHandle{"client", 0}

	// ========================================= //
	//*
	s1, err := tcp.NewServer(":8081", hub, srvhandle, defaultEncryptionWord)
	if err != nil {
		log.Fatalln(err)
	}
	defer s1.Stop()
	go s1.Serve()
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		_, err := tcp.NewClient("127.0.0.1:8081", clihandle, defaultEncryptionWord, false)
		if err != nil {
			log.Fatalln(err)
		}
	}
	/*/
	// ========================================= //
	// s2, err := ws.NewServer(":8080", "/", hub, srvhandle)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer s2.Stop()

	// go s2.Serve()
	// for i := 0; i < 10; i++ {
	// 	_, err := ws.NewClient("127.0.0.1:8080", "/", clihandle, false)
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// }
	//defer s21.Close()*/
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

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
		log.Println(srvhandle.count / 30.0)
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
	log.Println(m.name, "OnConnect")
	c.SendPacket([]byte("ok === ok, " + m.name))
}
func (m *myHandle) OnDisconnect(c gnet.Connector) {
	log.Println(m.name, "OnDisconnect")
}

func (m *myHandle) OnPacket(c gnet.Connector, buf []byte) error {
	// msg := string(buf)
	// log.Println(m.name, "OnPacket", msg)
	atomic.AddInt32(&m.count, 1)
	c.SendPacket(buf)
	return nil
}
