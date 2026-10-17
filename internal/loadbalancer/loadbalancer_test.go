package loadbalancer_test

import (
	"net"
	"testing"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
	"github.com/anyangateny1/Load-Balancer/internal/loadbalancer"
)

func startLoadBalancer(t *testing.T, num_of_servers uint16, algo loadbalancer.BackendSelector) *loadbalancer.LoadBalancer {
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

	response := sendMessage(t, lb.Addr(), "HELLO\n")
	expected := "ACK: HELLO"

	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}
}

func TestRoundRobin(t *testing.T) {
	const numOfServers = 10
	lb := startLoadBalancer(t, numOfServers, &algorithm.RoundRobin{})

	for i := range numOfServers {
		response := sendMessage(t, lb.Addr(), "HELLO\n")
		expected := "ACK: HELLO"
		if response != expected {
			t.Errorf("request %d unexpected response: %q, want %q", i, response, expected)
		}
	}
}

func TestLoadBalancerFailsWithTooManyServers(t *testing.T) {
	const numOfServers = loadbalancer.MaxNumberOfServers + 1
	_, err := loadbalancer.NewLoadBalancer(numOfServers, nil)

	if err == nil {
		t.Fatalf("Expected LoadBalancer to fail but it succeeded")
	}
}

func TestLoadBalancerFailsWithZeroServers(t *testing.T) {
	const numOfServers = 0
	_, err := loadbalancer.NewLoadBalancer(numOfServers, nil)

	if err == nil {
		t.Fatalf("Expected LoadBalancer to fail but it succeeded")
	}
}
