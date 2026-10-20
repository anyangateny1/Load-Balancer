package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
	"github.com/anyangateny1/Load-Balancer/internal/loadbalancer"
)

const (
	defaultNumServers = 3
	defaultAlgo       = "roundrobin"
)

func main() {
	logger := slog.Default()

	algoFlag := flag.String("algo", defaultAlgo, "load balancing algorithm (roundrobin, random)")
	serversFlag := flag.Int("servers", defaultNumServers, "number of backend servers")
	flag.Parse()

	algo, err := parseAlgorithm(*algoFlag)
	if err != nil {
		log.Fatal(err)
	}

	var addrs []loadbalancer.Backend
	var backends []*backendserver.BackendServer

	for i := 0; i < *serversFlag; i++ {
		srv := backendserver.NewBackendServer(logger)
		if err := srv.StartListening(); err != nil {
			log.Fatalf("failed to start backend server %d: %v", i, err)
		}

		backends = append(backends, srv)
		addrs = append(addrs, srv.Addr())
	}

	lb, err := loadbalancer.NewLoadBalancer(addrs, algo, loadbalancer.NetworkConfig{})
	if err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")

		_ = lb.Close()

		for _, srv := range backends {
			_ = srv.Close()
		}
	}()

	fmt.Printf("load balancer listening on %s (algo=%s, backends=%d)\n",
		lb.Addr(), *algoFlag, *serversFlag)

	lb.AcceptConnections()
}

func parseAlgorithm(name string) (loadbalancer.BackendSelector, error) {
	switch name {
	case "roundrobin":
		return &algorithm.RoundRobin{}, nil
	case "random":
		return algorithm.NewRandom(), nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %q (options: roundrobin, random)", name)
	}
}
