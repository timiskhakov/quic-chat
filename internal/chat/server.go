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
	"log"
	"math/big"
	"sync"
	"time"
)

type server struct {
	listener quic.Listener
	clients  map[string]quic.Connection
	messages chan Message
	mutex    sync.Mutex
}

func NewServer(addr string) (*server, error) {
	tlsConf, err := generateTLSConfig()
	if err != nil {
		return nil, err
	}

	listener, err := quic.ListenAddr(addr, tlsConf, &quic.Config{
		KeepAlivePeriod: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &server{
		listener: listener,
		clients:  map[string]quic.Connection{},
		messages: make(chan Message),
	}, nil
}

func (s *server) Close() error {
	close(s.messages)
	return s.listener.Close()
}

func (s *server) Broadcast(ctx context.Context) {
	for {
		select {
		case message := <-s.messages:
			s.mutex.Lock()
			for _, client := range s.clients {
				stream, err := client.OpenStreamSync(context.Background())
				if err != nil {
					log.Printf("[ERROR] failed to open client stream: %v\n", err)
					continue
				}
				if err := gob.NewEncoder(stream).Encode(&message); err != nil {
					log.Printf("[ERROR] failed to decode message: %v\n", err)
					continue
				}
				_ = stream.Close()
			}
			s.mutex.Unlock()
			log.Printf("[INFO] message received, broadcasted to %d clients\n", len(s.clients))
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) Accept(ctx context.Context) {
	for {
		conn, err := s.listener.Accept(ctx)
		if err != nil {
			return
		}

		go s.handleConn(conn)
	}
}

func (s *server) handleConn(conn quic.Connection) {
	defer func() { _ = conn.CloseWithError(1, "server closed") }()

	s.mutex.Lock()
	s.clients[conn.RemoteAddr().String()] = conn
	s.mutex.Unlock()

	log.Printf("[INFO] added client: %s\n", conn.RemoteAddr().String())

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			addr := conn.RemoteAddr().String()
			s.mutex.Lock()
			if _, ok := s.clients[addr]; ok {
				delete(s.clients, addr)
				log.Printf("[INFO] removed client %s\n", addr)
			}
			s.mutex.Unlock()
			return
		}

		var message Message
		if err := gob.NewDecoder(stream).Decode(&message); err != nil {
			log.Printf("[ERROR] failed to decode message: %v\n", err)
			return
		}

		s.messages <- message

		_ = stream.Close()
	}
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
		NextProtos:   []string{protocol},
	}, nil
}
