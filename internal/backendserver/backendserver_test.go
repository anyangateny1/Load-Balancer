package backendserver_test

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
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

func TestClientFragmentation(t *testing.T) {
	server := startServer(t, 1)

	conn, err := net.Dial(server.Addr().Network(), server.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	fullMsg := "Hello Fragmented World\n"

	fragments := []string{
		"H",
		"ello ",
		"Frag",
		"mented ",
		"Wor",
		"ld",
		"\n",
	}

	for _, part := range fragments {
		if _, err := conn.Write([]byte(part)); err != nil {
			t.Fatalf("failed to write fragment: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // encourage packet splitting
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	response := string(buf[:n])
	expected := "Server 1 ACK: " + strings.ToUpper(fullMsg)

	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}
}

func TestMixedClientSpeeds_FastFinishesFirst(t *testing.T) {
	server := startServer(t, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	fastDone := make(chan time.Duration, 1)

	go func() {
		defer wg.Done()

		conn, err := net.Dial(server.Addr().Network(), server.Addr().String())
		if err != nil {
			t.Errorf("slow dial error: %v", err)
			return
		}
		defer conn.Close()

		fragments := []string{"H", "ello ", "World", "\n"}
		for _, part := range fragments {
			conn.Write([]byte(part))
			time.Sleep(100 * time.Millisecond) // very slow sender
		}

		buf := make([]byte, 1024)
		conn.Read(buf) // ignore result; slow client is just pressure
	}()

	go func() {
		defer wg.Done()

		start := time.Now()

		conn, err := net.Dial(server.Addr().Network(), server.Addr().String())
		if err != nil {
			t.Errorf("fast dial error: %v", err)
			return
		}
		defer conn.Close()

		msg := "FAST\n"
		conn.Write([]byte(msg))

		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			t.Errorf("fast read error: %v", err)
			return
		}

		fastDone <- time.Since(start)
	}()

	wg.Wait()

	duration := <-fastDone

	if duration > 150*time.Millisecond {
		t.Fatalf("fast client was delayed too long: %v ", duration)
	}
}
