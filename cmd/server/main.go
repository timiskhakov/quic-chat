package main

import (
	"fmt"
	"github.com/timiskhakov/quic-chat/internal/chat"
	"os"
)

const addr = "localhost:4242"

func main() {
	if err := run(); err != nil {
		fmt.Printf("unhandled application error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	server, closeSrv, err := chat.NewServer(addr)
	if err != nil {
		return err
	}
	defer closeSrv()

	go server.Accept()

	server.Broadcast()

	return nil
}
