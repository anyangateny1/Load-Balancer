package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
	"github.com/anyangateny1/Load-Balancer/internal/loadbalancer"
)

const (
	DEFAULT_NUMBER_OF_SERVERS             int    = 3
	DEFAULT_LOADBALANCER_BACKEND_SELECTOR string = "roundrobin"
)

func main() {
	algoFlag := flag.String("algo", DEFAULT_LOADBALANCER_BACKEND_SELECTOR, "load balancing algorithm (roundrobin, random)")
	servers := flag.Int("servers", DEFAULT_NUMBER_OF_SERVERS, "number of backend servers")
	flag.Parse()

	algo, err := parseAlgorithm(*algoFlag)
	if err != nil {
		log.Fatal(err)
	}

	lb, err := loadbalancer.NewLoadBalancer(uint16(*servers), algo)
	if err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")
		_ = lb.Close()
	}()

	fmt.Printf("load balancer listening on %s (algo=%s, backends=%d)\n",
		lb.Addr(), *algoFlag, *servers)
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
