package udpmux

import (
	"fmt"
	"log"
	"net"
)

type Cfg struct {
	Outputs    map[string]interface{}
	Port       string
	Debug      bool
	BufferSize int
}

func Start(cfg *Cfg, l *log.Logger) {
	if cfg.Debug {
		l.Printf("%s", cfg)
	}

	myAddress, err := net.ResolveUDPAddr("udp", "localhost:"+cfg.Port)
	ln, err := net.ListenUDP("udp", myAddress)
	if err != nil {
		l.Println(err)
		return
	}

	l.Println("Listening on", myAddress)
	buf := make([]byte, cfg.BufferSize)
	for {
		n, addr, err := ln.ReadFromUDP(buf)
		if err != nil {
			l.Println(err)
		}
		fmt.Println(addr, "sent", string(buf[0:n]))
	}
}
