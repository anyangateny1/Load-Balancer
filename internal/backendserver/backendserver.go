package backendserver

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
	"time"
)

type Handler interface {
	HandleMessage([]byte) ([]byte, error)
}
type BackendServer struct {
	logger   *slog.Logger
	listener net.Listener
	handler  Handler
}

func NewBackendServer(logger *slog.Logger, handler Handler) (*BackendServer, error) {
	if handler == nil {
		return nil, errors.New("backend: handler cannot be nil")
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &BackendServer{
		logger:  logger,
		handler: handler,
	}, nil
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

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		b.logger.Error("set read deadline", "error", err)
		return
	}

	reader := bufio.NewReader(conn)
	message, err := reader.ReadBytes('\n')
	if err != nil {
		b.logger.Error("Read Error", "error", err)
		return
	}

	ackMsg, err := b.handler.HandleMessage(message)
	if err != nil {
		b.logger.Error("Handle Message Error", "error", err)
		return
	}

	_, err = conn.Write(ackMsg)
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

func (b *BackendServer) Addr() net.Addr {
	if b.listener == nil {
		return nil
	}
	return b.listener.Addr()
}
