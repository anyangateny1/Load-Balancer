package backendserver_test

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
)

type EchoHandler struct{}

func (EchoHandler) HandleMessage(msg []byte) ([]byte, error) {
	return msg, nil
}

func startServer(t *testing.T) *backendserver.BackendServer {
	t.Helper()

	server, err := backendserver.NewBackendServer(slog.Default(), EchoHandler{})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if err := server.StartListening(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	t.Cleanup(func() {
		_ = server.Close()
	})

	return server
}

func sendMessage(t *testing.T, addr net.Addr, msg string) string {
	t.Helper()

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(msg)); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	return string(buf[:n])
}

func TestHandleConnection(t *testing.T) {
	server := startServer(t)

	msg := "Hello\n"

	response := sendMessage(t, server.Addr(), msg)

	if response != msg {
		t.Fatalf("unexpected response: %q, want %q", response, msg)
	}
}

func TestMultipleConnections(t *testing.T) {
	server := startServer(t)

	const numConnections = 5

	var wg sync.WaitGroup
	wg.Add(numConnections)

	for i := range numConnections {
		go func(id int) {
			defer wg.Done()

			msg := fmt.Sprintf("Hello from connection %d\n", id)
			response := sendMessage(t, server.Addr(), msg)

			if response != msg {
				t.Errorf(
					"connection %d: unexpected response: %q, want %q",
					id,
					response,
					msg,
				)
			}
		}(i)
	}

	wg.Wait()
}

func TestClientFragmentation(t *testing.T) {
	server := startServer(t)

	conn, err := net.Dial(server.Addr().Network(), server.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	fragments := []string{
		"H",
		"ello ",
		"Frag",
		"mented ",
		"Wor",
		"ld",
		"\n",
	}

	expected := "Hello Fragmented World\n"

	for _, fragment := range fragments {
		if _, err := conn.Write([]byte(fragment)); err != nil {
			t.Fatalf("failed to write fragment: %v", err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	response := string(buf[:n])

	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}
}

func TestMultipleClientsAreIndependent(t *testing.T) {
	server := startServer(t)

	var wg sync.WaitGroup
	wg.Add(2)

	fastDone := make(chan time.Duration, 1)

	go func() {
		defer wg.Done()

		conn, err := net.Dial(server.Addr().Network(), server.Addr().String())
		if err != nil {
			t.Errorf("slow client dial error: %v", err)
			return
		}
		defer conn.Close()

		fragments := []string{
			"H",
			"ello ",
			"World",
			"\n",
		}

		for _, fragment := range fragments {
			if _, err := conn.Write([]byte(fragment)); err != nil {
				t.Errorf("slow client write error: %v", err)
				return
			}

			time.Sleep(100 * time.Millisecond)
		}

		buf := make([]byte, 1024)
		if _, err := conn.Read(buf); err != nil {
			t.Errorf("slow client read error: %v", err)
		}
	}()

	go func() {
		defer wg.Done()

		start := time.Now()

		conn, err := net.Dial(server.Addr().Network(), server.Addr().String())
		if err != nil {
			t.Errorf("fast client dial error: %v", err)
			return
		}
		defer conn.Close()

		msg := "FAST\n"

		if _, err := conn.Write([]byte(msg)); err != nil {
			t.Errorf("fast client write error: %v", err)
			return
		}

		buf := make([]byte, 1024)

		n, err := conn.Read(buf)
		if err != nil {
			t.Errorf("fast client read error: %v", err)
			return
		}

		if response := string(buf[:n]); response != msg {
			t.Errorf("fast client response: %q, want %q", response, msg)
			return
		}

		fastDone <- time.Since(start)
	}()

	wg.Wait()

	duration := <-fastDone

	if duration > 150*time.Millisecond {
		t.Fatalf("fast client was delayed too long: %v", duration)
	}
}
