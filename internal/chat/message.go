package chat

import (
	"encoding/gob"
	"io"
)

type Message struct {
	Nickname string
	Text     string
}

func (m *Message) Read(stream io.Reader) error {
	return gob.NewDecoder(stream).Decode(m)
}

func (m *Message) Write(stream io.Writer) error {
	return gob.NewEncoder(stream).Encode(m)
}
