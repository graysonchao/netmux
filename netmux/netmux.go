package netmux

import (
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

func createOutputs(l *log.Logger) []Output {
	outputs := []Output{}

	for name := range viper.GetStringMap("outputs") {
		outputDef := viper.GetStringMapString("outputs." + name)
		w := Worker{
			in:   make(chan []byte, 128),
			quit: make(chan bool),
		}
		if outputDef["type"] == "pipe" {
			if err := mkfifo(outputDef["path"], 0660); err != nil {
				l.Printf("Couldn't create named pipe for output %s at %s: %s\n", name, outputDef["path"], err)
			}
			outPipe, err := os.OpenFile(outputDef["path"], os.O_RDWR, os.ModeNamedPipe)
			if err != nil {
				l.Printf("Couldn't open named pipe for output %s at %s: %s\n", name, outputDef["path"], err)
				continue
			}
			o := &PipeOutput{
				Worker: w,
				out:    outPipe,
			}
			go o.read()
			outputs = append(outputs, o)
		} else {
			var outAddr net.Addr
			var err error
			switch {
			case outputDef["type"] == "udp":
				outAddr, err = net.ResolveUDPAddr("udp", outputDef["address"])
			case outputDef["type"] == "unix":
				outAddr, err = net.ResolveUnixAddr("unix", outputDef["address"])
			case outputDef["type"] == "tcp":
				outAddr, err = net.ResolveTCPAddr("tcp", outputDef["address"])
			default:
				l.Printf("Found invalid output definition at %s: invalid type\n", name)
			}
			if err != nil {
				l.Printf("Couldn't resolve address %s: %s\n", outputDef["address"], err)
				continue
			}
			conn, err := net.Dial(outAddr.Network(), outAddr.String())
			if err != nil {
				l.Printf("Couldn't dial address %s: %s\n", outputDef["address"], err)
				continue
			}
			o := &NetOutput{
				Worker: w,
				out:    outAddr,
				conn:   conn,
			}
			go o.read()
			outputs = append(outputs, o)
		}
	}
	return outputs
}

func StartTCP(l *log.Logger) {
	//debug := viper.GetBool("debug")
	// Cleanup outputs on SIGTERM
	outputs := createOutputs(l)
	l.Printf("Created outputs %s", outputs)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, unix.SIGTERM)
	go func() {
		<-c
		for _, o := range outputs {
			o.teardown()
		}
		os.Exit(0)
	}()

	// Start listening for connections
	myAddress, err := net.ResolveTCPAddr("tcp", "localhost:"+viper.GetString("port"))
	ln, err := net.ListenTCP("tcp", myAddress)
	if err != nil {
		l.Printf("Couldn't start listening: %s\n", err)
		return
	}
	l.Println("Listening for TCP connections on", myAddress)

	// Main loop - accept a connection and handle it until it's closed.
	// We only accept a single TCP connection at a time.
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			l.Printf("Error accepting TCP connection: %s", err)
			conn.Close()
			return
		}

		receiveBufferSizeBytes := viper.GetInt("receiveBufferSizeBytes")
		buf := make([]byte, receiveBufferSizeBytes)
		bytesRead, err := conn.Read(buf)
		for err == nil {
			for _, o := range outputs {
				o.send(buf[0:bytesRead])
			}
			bytesRead, err = conn.Read(buf)
		}
		conn.Close()
	}
}

func StartUDP(l *log.Logger) {
	outputs := createOutputs(l)
	l.Printf("Created outputs %s", outputs)
	// Cleanup outputs on SIGTERM
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, unix.SIGTERM)
	go func() {
		<-c
		for _, o := range outputs {
			o.teardown()
		}
		os.Exit(0)
	}()

	receiveBufferSizeBytes := viper.GetInt("receiveBufferSizeBytes")
	buf := make([]byte, receiveBufferSizeBytes)

	myAddress, err := net.ResolveUDPAddr("udp", "localhost:"+viper.GetString("port"))
	ln, err := net.ListenUDP("udp", myAddress)
	if err != nil {
		l.Printf("Couldn't start listening: %s\n", err)
		return
	}
	l.Println("Listening for UDP connections on", myAddress)

	for {
		n, _, err := ln.ReadFromUDP(buf)
		if err != nil {
			l.Println(err)
		}
		for _, o := range outputs {
			o.send(buf[0:n])
		}
	}
}

func Start(l *log.Logger) {
	switch {
	case viper.GetString("input") == "tcp":
		StartTCP(l)
	case viper.GetString("input") == "udp":
		StartUDP(l)
	default:
		l.Printf("Invalid input type: %s", viper.GetString("input"))
	}
}
