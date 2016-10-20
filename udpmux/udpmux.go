package udpmux

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
		switch {
		case outputDef["type"] == "pipe":
			if err := mkfifo(outputDef["path"], 0660); err != nil {
				l.Printf("Couldn't create named pipe for output %s at %s: %s\n", name, outputDef["path"], err)
				continue
			}
			outPipe, err := os.OpenFile(outputDef["path"], os.O_RDWR, os.ModeNamedPipe)
			if err != nil {
				l.Printf("Couldn't open named pipe for output %s at %s: %s\n", name, outputDef["path"], err)
				continue
			}
			o := &PipeOutput{
				in:   make(chan []byte, 128),
				quit: make(chan bool),
				out:  outPipe,
			}
			go o.read()
			outputs = append(outputs, o)
		case outputDef["type"] == "udp":
			outAddr, err := net.ResolveUDPAddr("udp", outputDef["address"])
			if err != nil {
				l.Printf("Couldn't resolve address %s: %s\n", outputDef["address"], err)
				continue
			}
			o := &UDPOutput{
				in:   make(chan []byte, 128),
				quit: make(chan bool),
				out:  outAddr,
			}
			go o.read()
			outputs = append(outputs, o)
		case outputDef["type"] == "unix":
			outAddr, err := net.ResolveUnixAddr("unix", outputDef["address"])
			if err != nil {
				l.Printf("Couldn't resolve address %s: %s\n", outputDef["address"], err)
				continue
			}
			o := &UnixOutput{
				in:   make(chan []byte, 128),
				quit: make(chan bool),
				out:  outAddr,
			}
			go o.read()
			outputs = append(outputs, o)
		default:
			l.Printf("Found invalid output definition at %s: invalid type\n", name)
		}
	}
	return outputs
}

func Start(l *log.Logger) {
	//debug := viper.GetBool("debug")
	myAddress, err := net.ResolveUDPAddr("udp", "localhost:"+viper.GetString("port"))
	ln, err := net.ListenUDP("udp", myAddress)
	if err != nil {
		l.Printf("Couldn't start listening: %s\n", err)
		return
	}

	l.Println("Listening on", myAddress)
	receiveBufferSizeBytes := viper.GetInt("receiveBufferSizeBytes")
	buf := make([]byte, receiveBufferSizeBytes)

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
