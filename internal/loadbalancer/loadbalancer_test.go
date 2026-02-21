package loadbalancer_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
	"github.com/anyangateny1/Load-Balancer/internal/loadbalancer"
)

func startLoadBalancer(t *testing.T, num_of_servers int, algo algorithm.Algorithm) *loadbalancer.LoadBalancer {
	t.Helper()

	lb, err := loadbalancer.NewLoadBalancer(num_of_servers, algo)
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

	lb := startLoadBalancer(t, 10, &algorithm.RoundRobin{})

	response := sendMessage(t, lb.Addr(), "Hello\n")
	expected := "Server 0 ACK: HELLO\n"

	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}

}
func TestRoundRobin(t *testing.T) {

	const numOfServers = 10
	lb := startLoadBalancer(t, numOfServers, &algorithm.RoundRobin{})

	for i := 0; i < numOfServers; i++ {
		response := sendMessage(t, lb.Addr(), "HELLO\n")
		expected := fmt.Sprintf("Server %d ACK: HELLO\n", i)
		if response != expected {
			t.Errorf("request %d unexpected response: %q, want %q", i, response, expected)
		}
	}

}
