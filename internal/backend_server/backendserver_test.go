package backendserver_test

import (
	backendserver "load-balancer/internal/backend_server"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHandlingConnection(t *testing.T) {
	server, err := backendserver.NewBackendServer(1)

	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go server.AcceptConnections()
	defer func() { _ = server.Close() }()

	time.Sleep(100 * time.Millisecond)

	addr := server.Addr()
	network_type := addr.Network()

	conn, err := net.Dial(network_type, addr.String())
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

func TestMultipleConnections(t *testing.T) {
	server, err := backendserver.NewBackendServer(1)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go server.AcceptConnections()
	defer func() { _ = server.Close() }()

	time.Sleep(100 * time.Millisecond)

	addr := server.Addr()
	networkType := addr.Network()

	numConnections := 5
	var wg sync.WaitGroup
	wg.Add(numConnections)

	for i := range numConnections {
		go func(connID int) {
			defer wg.Done()

			conn, err := net.Dial(networkType, addr.String())
			if err != nil {
				t.Errorf("connection %d failed: %v", connID, err)
				return
			}
			defer func() { _ = server.Close() }()

			msg := fmt.Sprintf("Hello from connection %d\n", connID)
			_, err = conn.Write([]byte(msg))
			if err != nil {
				t.Errorf("connection %d failed to write: %v", connID, err)
				return
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				t.Errorf("connection %d failed to read: %v", connID, err)
				return
			}

			response := string(buf[:n])
			expected := fmt.Sprintf("ACK: %s", strings.ToUpper(msg))
			if response != expected {
				t.Errorf("connection %d unexpected response: %q, want %q", connID, response, expected)
			}
		}(i)
	}

	wg.Wait()
}
