package backendserver_test

import (
	"fmt"
	"load-balancer/internal/backendserver"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func startServer(t *testing.T, id int) *backendserver.BackendServer {
	t.Helper()

	server, err := backendserver.NewBackendServer(id)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go server.AcceptConnections()
	t.Cleanup(func() { _ = server.Close() })

	time.Sleep(100 * time.Millisecond)
	return server
}

func sendMessage(t *testing.T, addr net.Addr, msg string) string {
	t.Helper()

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	return string(buf[:n])
}

func TestHandlingConnection(t *testing.T) {
	server := startServer(t, 1)

	response := sendMessage(t, server.Addr(), "Hello\n")
	expected := "Server 1 ACK: HELLO\n"
	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}
}

func TestMultipleConnections(t *testing.T) {
	server := startServer(t, 1)

	numConnections := 5
	var wg sync.WaitGroup
	wg.Add(numConnections)

	for i := range numConnections {
		go func(connID int) {
			defer wg.Done()

			msg := fmt.Sprintf("Hello from connection %d\n", connID)
			response := sendMessage(t, server.Addr(), msg)
			expected := fmt.Sprintf("Server 1 ACK: %s", strings.ToUpper(msg))
			if response != expected {
				t.Errorf("connection %d unexpected response: %q, want %q", connID, response, expected)
			}
		}(i)
	}

	wg.Wait()
}
