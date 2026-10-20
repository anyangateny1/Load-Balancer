// Package loadbalancer implements a Layer 4 TCP load balancer.
package loadbalancer

import (
	"errors"
	"io"
	"log/slog"
	"net"
)

// MaxNumberOfServers is the maximum number of backend servers supported.
const (
	MaxNumberOfServers uint16 = 1000
	defaultAddress     string = "localhost:8080"
)

// Backend is the interface required for a load balancer server.
type Backend interface {
	net.Addr
}

// BackendSelector selects the index of the next backend server.
type BackendSelector interface {
	Next(numBackends int) int
}

// LoadBalancer is a Layer 4 load balancer that pipes connections to servers.
// It selects the `next` server via an injected algorithm.
type LoadBalancer struct {
	algo     BackendSelector
	backends []Backend
	logger   *slog.Logger
	listener net.Listener
}

// NetworkConfig contains network configuration for the load balancer, including port and address.
type NetworkConfig struct {
	Endpoint string
}

// NewLoadBalancer creates a load balancer with the specified number of backend
// servers, selection algorithm, and network configuration.
func NewLoadBalancer(servers []Backend, algo BackendSelector, cfg NetworkConfig) (*LoadBalancer, error) {
	numOfServers := len(servers)
	if numOfServers < 1 || numOfServers > int(MaxNumberOfServers) {
		return nil, errors.New("number of servers must be between 1 and 1000")
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = defaultAddress
	}

	logger := slog.Default()

	ln, err := net.Listen("tcp", endpoint)
	if err != nil {
		return nil, err
	}

	return &LoadBalancer{
		algo:     algo,
		backends: servers,
		logger:   logger,
		listener: ln,
	}, nil
}

// TODO: refactor and make AcceptConnections return the connection.

// AcceptConnections accepts incoming client connections and proxies them to backend servers concurrently.
func (lb *LoadBalancer) AcceptConnections() {
	for {
		conn, err := lb.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				lb.logger.Info("LoadBalancer gracefully shutting down")
				return
			}
			lb.logger.Error("LoadBalancer Error while accepting", "error", err)
			return
		}
		go lb.pipeConnections(conn)
	}
}

func (lb *LoadBalancer) pipeConnections(clientConn net.Conn) {
	index := lb.algo.Next(len(lb.backends))
	backendAddr := lb.backends[index]

	backendConn, err := net.Dial(backendAddr.Network(), backendAddr.String())
	if err != nil {
		lb.logger.Error("Failed to connect to backend:", "error", err, "server", index)
		_ = clientConn.Close()
		return
	}

	go func() {
		_, _ = io.Copy(backendConn, clientConn)

		if tcp, ok := backendConn.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
	}()

	go func() {
		_, _ = io.Copy(clientConn, backendConn)

		_ = clientConn.Close()
		_ = backendConn.Close()
	}()
}

// Close closes the backend servers and the load balancer's listener.
func (lb *LoadBalancer) Close() error {
	return lb.listener.Close()
}

// Addr returns the load balancer's listening address.
func (lb *LoadBalancer) Addr() net.Addr {
	if lb.listener != nil {
		return lb.listener.Addr()
	}

	return nil
}
