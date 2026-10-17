// Package loadbalancer implements a Layer 4 TCP load balancer.
package loadbalancer

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
)

// TODO: Look into injecting port?

// MaxNumberOfServers is the maximum number of backend servers supported.
const MaxNumberOfServers uint16 = 1000

// BackendSelector selects the index of the next backend server.
type BackendSelector interface {
	Next(numBackends int) int
}

// LoadBalancer is a Layer 4 load balancer that pipes connections to servers.
// It selects the `next` server via an injected algorithm.
type LoadBalancer struct {
	algo     BackendSelector
	backends []*backendserver.BackendServer
	logger   *slog.Logger
	listener net.Listener
}

// TODO: Refactor to accept a list of addresses

// TODO: Refactor to accept a port

// NewLoadBalancer creates a load balancer with the specified number of backend
// servers and selection algorithm.
func NewLoadBalancer(numOfServers uint16, algo BackendSelector) (*LoadBalancer, error) {
	if numOfServers < 1 || numOfServers > MaxNumberOfServers {
		return nil, errors.New("number of servers must be between 1 and 1000")
	}

	var servers []*backendserver.BackendServer
	logger := slog.Default()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	for i := range numOfServers {
		serv := backendserver.NewBackendServer(logger)

		if err := serv.StartListening(); err != nil {
			_ = ln.Close()
			return nil, fmt.Errorf("failed to start backend server %d: %w", i, err)
		}

		logger.Info("server started", "server", i, "addr", serv.Addr())

		servers = append(servers, serv)
	}

	if len(servers) == 0 {
		return nil, errors.New("failed to start any backend servers")
	}

	if uint16(len(servers)) < numOfServers {
		logger.Error(
			"Started only servers",
			"started", len(servers),
			"expected", numOfServers,
		)
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
	backendAddr := lb.backends[index].Addr()

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
	for _, b := range lb.backends {
		err := b.Close()
		if err != nil {
			lb.logger.Error("Failed to close backend:", "error", err, "server", b.Addr().String())
			continue
		}
	}

	return lb.listener.Close()
}

// Addr returns the load balancer's listening address.
func (lb *LoadBalancer) Addr() net.Addr {
	if lb.listener != nil {
		return lb.listener.Addr()
	}

	return nil
}
