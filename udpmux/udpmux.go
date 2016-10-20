package udpmux

import (
	"log"
	"net"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

func createOutputs(l *log.Logger) []Output {
	outputs := []Output{}

	for name := range viper.GetStringMap("outputs") {
		outputDef := viper.GetStringMapString("outputs." + name)
		switch {
		case outputDef["type"] == "pipe":
			if err := unix.Mkfifo(outputDef["path"], 0666); err != nil {
				l.Printf("Couldn't create named pipe for output %s at %s: %s\n", name, outputDef["path"], err)
			}
			outPipe, err := os.OpenFile(outputDef["path"], os.O_RDWR, os.ModeNamedPipe)
			if err != nil {
				l.Printf("Couldn't open named pipe for output %s at %s: %s\n", name, outputDef["path"], err)
			}
			o := &PipeOutput{
				in:   make(chan []byte, 128),
				quit: make(chan bool),
				out:  outPipe,
			}
			go o.read()
			//output := NewPipeOutput(outputDef)
			outputs = append(outputs, o)
		case outputDef["type"] == "remote":
			outAddr, err := net.ResolveUDPAddr("udp", outputDef["address"])
			if err != nil {
				l.Printf("Couldn't resolve address %s: %s\n", outputDef["address"], err)
			}
			o := &RemoteOutput{
				in:   make(chan []byte, 128),
				quit: make(chan bool),
				out:  outAddr,
			}
			go o.read()
			//output := NewRemoteOutput(outputDef)
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
