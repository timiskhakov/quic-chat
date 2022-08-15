package chat

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"github.com/lucas-clemente/quic-go"
	"math/big"
	"sync"
)

type server struct {
	listener quic.Listener
	conns    map[string]quic.Connection
	messages chan Message
	mutex    sync.Mutex
}

func NewServer(addr string) (*server, func(), error) {
	tlsConf, err := generateTLSConfig()
	if err != nil {
		return nil, func() {}, err
	}

	listener, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		return nil, func() {}, err
	}

	messages := make(chan Message)

	return &server{
			listener: listener,
			conns:    map[string]quic.Connection{},
			messages: messages,
		}, func() {
			close(messages)
			_ = listener.Close()
		}, nil
}

func (s *server) Broadcast() {
	for message := range s.messages {
		s.mutex.Lock()
		for _, conn := range s.conns {
			stream, err := conn.OpenStreamSync(context.Background())
			if err != nil {
				// TODO: Handle error
				continue
			}
			if err := gob.NewEncoder(stream).Encode(&message); err != nil {
				// TODO: Handle error
				continue
			}
			_ = stream.Close()
		}
		s.mutex.Unlock()
	}
}

func (s *server) Accept() {
	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			// TODO: Handle error
			continue
		}

		go s.handleConn(conn)
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

	var message Message
	if err := gob.NewDecoder(stream).Decode(&message); err != nil {
		// TODO: handle error
		return
	}

	s.messages <- message
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}, nil
}
