package main

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucas-clemente/quic-go"
	"github.com/timiskhakov/quic-chat/internal"
	"log"
)

const addr = "localhost:4242"

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	conn, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		// TODO: Handle error properly
		return
	}
	defer func() { _ = conn.CloseWithError(1, "Out!") }()

	ch := make(chan internal.Message)
	defer close(ch)

	m := initialModel(func(text string) {
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			return
		}
		defer func() { _ = stream.Close() }()

		msg := internal.Message{Nickname: "Tim", Text: text}

		_ = gob.NewEncoder(stream).Encode(&msg)
	}, ch)

	// Chat

	go func() {
		for {
			stream, err := conn.AcceptStream(context.Background())

			if err != nil {
				fmt.Println(err)
				return
			}

			var message internal.Message
			if err := gob.NewDecoder(stream).Decode(&message); err != nil {
				return
			}

			ch <- message

			_ = stream.Close()
		}
	}()

	p := tea.NewProgram(m)

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
