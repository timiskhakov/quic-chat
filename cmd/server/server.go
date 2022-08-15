package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/timiskhakov/quic-chat/internal"
	"sync"
)

type server struct {
	mutex    sync.Mutex
	conns    map[string]quic.Connection
	messages chan internal.Message
}

func NewServer() *server {
	return &server{
		conns:    map[string]quic.Connection{},
		messages: make(chan internal.Message),
	}
}

func (s *server) Start() {
	defer close(s.messages)

	for {
		select {
		case message := <-s.messages:
			s.mutex.Lock()
			for _, conn := range s.conns {
				stream, err := conn.OpenStreamSync(context.Background())
				if err != nil {
					// TODO: Handle error
					continue
				}

				if err := gob.NewEncoder(stream).Encode(&message); err != nil {
					// TODO: Handle error
					return
				}
			}
			s.mutex.Unlock()
			fmt.Printf("Sent to %d users\n", len(s.conns))
		}
	}
}

func (s *server) handleConn(conn quic.Connection) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			s.mutex.Lock()
			delete(s.conns, conn.RemoteAddr().String())
			s.mutex.Unlock()
			// TODO: Add logging?
			return
		}

		s.mutex.Lock()
		s.conns[conn.RemoteAddr().String()] = conn
		s.mutex.Unlock()

		go s.handleStream(stream)
	}
}

func (s *server) handleStream(stream quic.Stream) {
	defer func() { _ = stream.Close() }()

	var message internal.Message
	if err := gob.NewDecoder(stream).Decode(&message); err != nil {
		// TODO: handle error
		return
	}

	s.messages <- message
}
