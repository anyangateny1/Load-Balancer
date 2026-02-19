package backendserver_test

import (
	backendserver "load-balancer/internal/backend_server"
	"net"
	"testing"
	"time"
)

func TestNewBackendServer(t *testing.T) {
	server := backendserver.NewBackendServer(1)

	if server.ServerNumber != 1 {
		t.Errorf("expected ServerNumber to be 1, got %d", server.ServerNumber)
	}
}

func TestHandlingConnection(t *testing.T) {
	server := backendserver.NewBackendServer(1)
	go server.ListenAndServe()
	defer func() { _ = server.Close() }()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Write([]byte("Hello\n"))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	response := string(buf[:n])
	expected := "ACK: HELLO\n"
	if response != expected {
		t.Fatalf("unexpected response: %s", response)
	}
}
