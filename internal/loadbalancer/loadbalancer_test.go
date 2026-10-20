package loadbalancer_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
	"github.com/anyangateny1/Load-Balancer/internal/loadbalancer"
)

func startLoadBalancer(t *testing.T, servers []loadbalancer.Backend, algo loadbalancer.BackendSelector) *loadbalancer.LoadBalancer {
	t.Helper()

	lb, err := loadbalancer.NewLoadBalancer(servers, algo)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go lb.AcceptConnections()
	t.Cleanup(func() { _ = lb.Close() })

	return lb
}

func startServer(t *testing.T, response string) loadbalancer.Backend {
	t.Helper()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_, _ = fmt.Fprint(conn, response)
			_ = conn.Close()
		}
	}()

	return ln.Addr()
}

func startServers(t *testing.T, count int) []loadbalancer.Backend {
	t.Helper()

	addrs := make([]loadbalancer.Backend, count)
	for i := range count {
		addrs[i] = startServer(t, fmt.Sprintf("Server %d", i))
	}
	return addrs
}

func sendMessage(t *testing.T, addr net.Addr, msg string) string {
	t.Helper()

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

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

func TestPacketForwarding(t *testing.T) {
	server := startServer(t, "ACK: HELLO")
	lb := startLoadBalancer(t, []loadbalancer.Backend{server}, &algorithm.RoundRobin{})

	response := sendMessage(t, lb.Addr(), "HELLO\n")
	expected := "ACK: HELLO"

	if response != expected {
		t.Fatalf("unexpected response: %q, want %q", response, expected)
	}
}

func TestRoundRobin(t *testing.T) {
	const numOfServers = 10
	servers := startServers(t, numOfServers)
	lb := startLoadBalancer(t, servers, &algorithm.RoundRobin{})

	for i := range numOfServers {
		response := sendMessage(t, lb.Addr(), "HELLO\n")
		expected := fmt.Sprintf("Server %d", i)
		if response != expected {
			t.Errorf("request %d unexpected response: %q, want %q", i, response, expected)
		}
	}
}

func TestLoadBalancerFailsWithTooManyServers(t *testing.T) {
	servers := make([]loadbalancer.Backend, loadbalancer.MaxNumberOfServers+1)
	_, err := loadbalancer.NewLoadBalancer(servers, &algorithm.RoundRobin{})

	if err == nil {
		t.Fatalf("Expected LoadBalancer to fail but it succeeded")
	}
}

func TestLoadBalancerFailsWithZeroServers(t *testing.T) {
	_, err := loadbalancer.NewLoadBalancer(nil, &algorithm.RoundRobin{})

	if err == nil {
		t.Fatalf("Expected LoadBalancer to fail but it succeeded")
	}
}
