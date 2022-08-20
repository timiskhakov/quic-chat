package main

import (
	"context"
	"fmt"
	"github.com/timiskhakov/quic-chat/internal/chat"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("unhandled application error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	server, err := chat.NewServer()
	if err != nil {
		return err
	}
	defer func() { _ = server.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Accept(ctx)
	go server.Broadcast(ctx)

	log.Println("server started")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs

	log.Println("shutting down server")

	return nil
}
