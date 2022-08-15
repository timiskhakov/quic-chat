package main

import (
	"fmt"
	"github.com/timiskhakov/quic-chat/internal/chat"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const addr = "localhost:4242"

func main() {
	if err := run(); err != nil {
		fmt.Printf("unhandled application error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	server, err := chat.NewServer(addr)
	if err != nil {
		return err
	}
	defer func() { _ = server.Close() }()

	go server.Accept()
	go server.Broadcast()

	log.Printf("server started: %s\n", addr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs

	log.Printf("shutting down server: %s\n", addr)

	return nil
}
