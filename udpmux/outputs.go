package udpmux

import (
	"fmt"
	"net"
	"os"
)

type PipeOutput struct {
	in   chan []byte
	out  *os.File
	quit chan bool
}

type RemoteOutput struct {
	in   chan []byte
	out  *net.UDPAddr
	quit chan bool
	conn *net.UDPConn
}

type Output interface {
	send([]byte)
}

func (o *PipeOutput) send(chunk []byte) {
	o.in <- chunk
}

func (o *RemoteOutput) send(chunk []byte) {
	o.in <- chunk
}

func (o *PipeOutput) read() {
	for {
		select {
		case chunk := <-o.in:
			if _, err := o.out.Write(chunk); err != nil {
				fmt.Println("Error writing to outfile! Terminating output")
				o.quit <- true
			}
		case <-o.quit:
			return
		}
	}
}

func (o *RemoteOutput) read() {
	conn, err := net.DialUDP("udp", nil, o.out)
	if err != nil {
		fmt.Printf("Couldn't dial address %s: %s\n", o.out, err)
	}
	o.conn = conn
	for {
		select {
		case chunk := <-o.in:
			if _, err := conn.Write(chunk); err != nil {
				conn, err := o.tryReconnect(chunk)
				if err != nil {
					fmt.Println("Error writing to remote addr! Terminating output")
					o.quit <- true
				}
				o.conn = conn
			}
		case <-o.quit:
			return
		}
	}
}

func (o *RemoteOutput) tryReconnect(chunk []byte) (*net.UDPConn, error) {
	var err error
	for i := 0; i < 5; i++ {
		conn, err := net.DialUDP("udp", nil, o.out)
		if err == nil {
			conn.Write(chunk)
			return conn, nil
		}
	}
	return nil, err
}
