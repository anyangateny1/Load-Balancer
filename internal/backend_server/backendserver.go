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
	ServerNumber int
	logger       *slog.Logger
	listener     net.Listener
}

func NewBackendServer(num int) *BackendServer {
	return &BackendServer{
		ServerNumber: num,
		logger:       slog.Default(),
	}
}

func (b *BackendServer) ListenAndServe() {
	var err error
	b.listener, err = net.Listen("tcp", ":8080")
	if err != nil {
		b.logger.Error("Listen Error", "error", err, "server", b.ServerNumber)
		return
	}

	for {
		conn, err := b.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				b.logger.Info("Listener closed, stopping server", "server", b.ServerNumber)
				return
			}
			b.logger.Error("Listener closed", "error", err, "server", b.ServerNumber)
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
		b.logger.Error("Read Error", "error", err, "server", b.ServerNumber)
		return
	}

	ackMsg := strings.ToUpper(strings.TrimSpace(message))
	response := fmt.Sprintf("ACK: %s\n", ackMsg)
	_, err = conn.Write([]byte(response))
	if err != nil {
		b.logger.Error("Server Write Error", "error", err, "server", b.ServerNumber)
	}
}

func (b *BackendServer) Close() error {
	if b.listener != nil {
		return b.listener.Close()
	}
	return nil
}
