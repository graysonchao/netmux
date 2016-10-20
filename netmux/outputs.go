package netmux

import (
	"fmt"
	"net"
	"os"
)

type Worker struct {
	in   chan []byte
	quit chan bool
}

type PipeOutput struct {
	Worker
	out *os.File
}

type NetOutput struct {
	Worker
	out  net.Addr
	conn net.Conn
}

type Output interface {
	send([]byte)
	teardown()
}

func (w *Worker) send(chunk []byte) {
	w.in <- chunk
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

func (o *PipeOutput) teardown() {}

func (n *NetOutput) read() {
	defer n.conn.Close()
	for {
		select {
		case chunk := <-n.in:
			if _, err := n.conn.Write(chunk); err != nil {
				conn, err := n.tryReconnect(chunk)
				if err != nil {
					fmt.Println("Error writing to remote addr! Terminating output")
					n.quit <- true
				}
				n.conn = conn
			}
		case <-n.quit:
			return
		}
	}
}

func (n *NetOutput) tryReconnect(chunk []byte) (net.Conn, error) {
	var err error
	for i := 0; i < 5; i++ {
		conn, err := net.Dial(n.out.Network(), n.out.String())
		if err == nil {
			conn.Write(chunk)
			return n.conn, nil
		}
	}
	return nil, err
}

func (n *NetOutput) teardown() {
	if n.conn != nil {
		n.conn.Close()
	}
}
