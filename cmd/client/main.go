package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/timiskhakov/quic-chat/internal/chat"
	"os"
)

const addr = "localhost:4242"

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	client, err := chat.NewClient(addr, "Tim")
	if err != nil {
		return err
	}

	m := createApp(client.Send, client.Receive())
	if err := tea.NewProgram(m).Start(); err != nil {
		return err
	}

	return nil
}
