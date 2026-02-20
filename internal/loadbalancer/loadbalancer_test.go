package loadbalancer_test

import (
	// "fmt"
	"github.com/anyangateny1/Load-Balancer/internal/loadbalancer"
	"net"
	// "strings"
	// "sync"
	"testing"
	// "time"
)

func startLoadBalancer(t *testing.T, num_of_servers int) *loadbalancer.LoadBalancer {
	t.Helper()

	lb, err := loadbalancer.NewLoadBalancer(num_of_servers)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go lb.AcceptConnections()
	t.Cleanup(func() { _ = lb.Close() })

	return lb
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

func TestPacketForwarding(t *testing.T) {

	lb := startLoadBalancer(t, 1)

	response := sendMessage(t, lb.Addr(), "Hello\n")
	expected := "Server 1 ACK: HELLO\n"

	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}

}
