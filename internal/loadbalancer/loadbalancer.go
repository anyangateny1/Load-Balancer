package loadbalancer

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
)

type LoadBalancer struct {
	algo     algorithm.Algorithm
	backend  []*backendserver.BackendServer
	logger   *slog.Logger
	listener net.Listener
}

func NewLoadBalancer(numOfServers int, algo algorithm.Algorithm) (*LoadBalancer, error) {

	const MaxServers = 1000
	if numOfServers == 0 || numOfServers > MaxServers {
		return nil, errors.New("number of servers must be between 1 and 1000")
	}

	var servers []*backendserver.BackendServer
	slog := slog.Default()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	for i := range numOfServers {
		var serv *backendserver.BackendServer
		var err error

		for attempt := 1; attempt <= 3; attempt++ {
			serv, err = startServer(i)
			if err == nil {
				break
			}

			slog.Error("server start failed",
				"server", i,
				"attempt", attempt,
				"error", err,
			)

			time.Sleep(time.Second)
		}

		if err != nil {
			slog.Error("server permanently failed to start", "server", i)
			continue
		}

		servers = append(servers, serv)
		go func(id int, s *backendserver.BackendServer) {
			slog.Info("server starting", "server", id)

			s.AcceptConnections()

			slog.Info("server stopped", "server", id)
		}(i, serv)
	}

	if len(servers) == 0 {
		return nil, errors.New("failed to start any backend servers")
	}

	return &LoadBalancer{
		algo:     algo,
		backend:  servers,
		logger:   slog,
		listener: ln,
	}, nil

}

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
	index := lb.algo.Next(len(lb.backend))
	backendAddr := lb.backend[index].Addr()

	backendConn, err := net.Dial(backendAddr.Network(), backendAddr.String())
	if err != nil {
		lb.logger.Error("Failed to connect to backend:", "error", err, "server", index)
		_ = clientConn.Close()
		return
	}

	go func() {
		defer func() { _ = backendConn.Close() }()
		defer func() { _ = clientConn.Close() }()
		_, _ = io.Copy(backendConn, clientConn)
	}()

	go func() {
		defer func() { _ = backendConn.Close() }()
		defer func() { _ = clientConn.Close() }()
		_, _ = io.Copy(clientConn, backendConn)
	}()
}

func (lb *LoadBalancer) Close() error {
	for _, b := range lb.backend {
		err := b.Close()
		if err != nil {
			lb.logger.Error("Failed to close backend:", "error", err, "server", b.Addr().String())
			continue
		}
	}

	return lb.listener.Close()
}

func (lb *LoadBalancer) Addr() net.Addr {
	if lb.listener != nil {
		return lb.listener.Addr()
	}

	return nil
}

func startServer(id int) (*backendserver.BackendServer, error) {
	return backendserver.NewBackendServer(id)
}
