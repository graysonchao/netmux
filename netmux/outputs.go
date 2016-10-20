package netmux

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

type UDPOutput struct {
	in   chan []byte
	out  *net.UDPAddr
	quit chan bool
	conn *net.UDPConn
}

type UnixOutput struct {
	in   chan []byte
	out  *net.UnixAddr
	quit chan bool
	conn *net.UnixConn
}

type TCPOutput struct {
	in   chan []byte
	out  *net.TCPAddr
	quit chan bool
	conn *net.TCPConn
}

type Output interface {
	send([]byte)
	teardown()
}

func (o *PipeOutput) send(chunk []byte) {
	o.in <- chunk
}

func (o *UDPOutput) send(chunk []byte) {
	o.in <- chunk
}

func (o *UnixOutput) send(chunk []byte) {
	o.in <- chunk
}

func (o *TCPOutput) send(chunk []byte) {
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
			o.out.Write([]byte("\n===\n"))
		case <-o.quit:
			return
		}
	}
}

func (o *PipeOutput) teardown() {
}

func (o *UDPOutput) read() {
	conn, err := net.DialUDP("udp", nil, o.out)
	if err != nil {
		fmt.Printf("Couldn't dial address %s: %s\n", o.out, err)
		return
	}
	o.conn = conn
	defer o.conn.Close()
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

func (o *UDPOutput) tryReconnect(chunk []byte) (*net.UDPConn, error) {
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

func (o *UDPOutput) teardown() {
	if o.conn != nil {
		o.conn.Close()
	}
}

func (o *UnixOutput) read() {
	conn, err := net.DialUnix("unix", nil, o.out)
	if err != nil {
		fmt.Printf("Couldn't dial address %s: %s\n", o.out, err)
		return
	}
	o.conn = conn
	defer o.conn.Close()
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

func (o *UnixOutput) tryReconnect(chunk []byte) (*net.UnixConn, error) {
	var err error
	for i := 0; i < 5; i++ {
		conn, err := net.DialUnix("udp", nil, o.out)
		if err == nil {
			conn.Write(chunk)
			return conn, nil
		}
	}
	return nil, err
}

func (o *UnixOutput) teardown() {
	if o.conn != nil {
		o.conn.Close()
	}
	os.Remove(o.out.String())
}

func (o *TCPOutput) read() {
	conn, err := net.DialTCP("tcp", nil, o.out)
	if err != nil {
		fmt.Printf("Couldn't dial address %s: %s\n", o.out, err)
		return
	}
	o.conn = conn
	defer o.conn.Close()
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

func (o *TCPOutput) tryReconnect(chunk []byte) (*net.TCPConn, error) {
	var err error
	for i := 0; i < 5; i++ {
		conn, err := net.DialTCP("udp", nil, o.out)
		if err == nil {
			conn.Write(chunk)
			return conn, nil
		}
	}
	return nil, err
}

func (o *TCPOutput) teardown() {
	if o.conn != nil {
		o.conn.Close()
	}
}
