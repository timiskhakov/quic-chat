package chat

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"github.com/lucas-clemente/quic-go"
)

type client struct {
	nickname string
	conn     quic.Connection
}

func NewClient(addr, nickname string) (*client, error) {
	conn, err := quic.DialAddr(addr, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{protocol},
	}, nil)
	if err != nil {
		return nil, err
	}

	return &client{nickname: nickname, conn: conn}, nil
}

func (c *client) Send(text string) error {
	stream, err := c.conn.OpenStream()
	if err != nil {
		return err
	}
	
	go c.writeStream(stream, text)

	return nil
}

func (c *client) Receive(ctx context.Context) (<-chan Message, <-chan error) {
	messages := make(chan Message)
	errs := make(chan error)
	go func() {
		defer close(messages)
		for {
			stream, err := c.conn.AcceptStream(ctx)
			if err != nil {
				errs <- err
				return
			}
			go c.readStream(stream, messages)
		}
	}()

	return messages, errs
}

func (c *client) writeStream(stream quic.Stream, text string) {
	defer func() { _ = stream.Close() }()

	message := Message{Nickname: c.nickname, Text: text}
	if err := gob.NewEncoder(stream).Encode(&message); err != nil {
		return
	}
}

func (c *client) readStream(stream quic.Stream, messages chan<- Message) {
	defer func() { _ = stream.Close() }()

	var message Message
	if err := gob.NewDecoder(stream).Decode(&message); err != nil {
		return
	}

	messages <- message
}
