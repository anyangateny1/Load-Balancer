package backendserver

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
)

type BackendServer struct {
	serverNumber int
	logger       *slog.Logger
	listener     net.Listener
}

func NewBackendServer(num int) (*BackendServer, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	return &BackendServer{
		serverNumber: num,
		logger:       slog.Default(),
		listener:     ln,
	}, nil
}

func (b *BackendServer) AcceptConnections() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				b.logger.Info("Listener closed, stopping server", "server", b.serverNumber)
				return
			}
			b.logger.Error("Error while accepting", "error", err, "server", b.serverNumber)
			return
		}
		go b.handleConnection(conn)
	}
}

func (b *BackendServer) handleConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		b.logger.Error("Read Error", "error", err, "server", b.serverNumber)
		return
	}

	ackMsg := strings.ToUpper(strings.TrimSpace(message))
	response := fmt.Sprintf("Server %d ACK: %s\n", b.serverNumber, ackMsg)
	_, err = conn.Write([]byte(response))
	if err != nil {
		b.logger.Error("Server Write Error", "error", err, "server", b.serverNumber)
	}
}

func (b *BackendServer) Close() error {
	if b.listener != nil {
		return b.listener.Close()
	}
	return nil
}

func (b *BackendServer) Addr() net.Addr {
	if b.listener == nil {
		return nil
	}
	return b.listener.Addr()
}
