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
	logger   *slog.Logger
	listener net.Listener
}

func NewBackendServer(logger *slog.Logger) *BackendServer {
	return &BackendServer{
		logger: logger,
	}
}

func (b *BackendServer) StartListening() error {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	b.listener = ln
	go b.AcceptConnections()

	return nil
}

func (b *BackendServer) AcceptConnections() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				b.logger.Info("Listener closed, stopping server")
				return
			}
			b.logger.Error("Error while accepting", "error", err)
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
		b.logger.Error("Read Error", "error", err)
		return
	}

	ackMsg := strings.ToUpper(strings.TrimSpace(message))
	response := fmt.Sprintf("ACK: %s\n", ackMsg)
	_, err = conn.Write([]byte(response))
	if err != nil {
		b.logger.Error("Server Write Error", "error", err)
	}
}

func (b *BackendServer) Close() error {
	if b.listener != nil {
		return b.listener.Close()
	}
	return nil
}

func (b *BackendServer) Conn() net.Listener {
	if b.listener != nil {
		return b.listener
	}
	return nil
}

func (b *BackendServer) Addr() net.Addr {
	if b.listener == nil {
		return nil
	}
	return b.listener.Addr()
}
