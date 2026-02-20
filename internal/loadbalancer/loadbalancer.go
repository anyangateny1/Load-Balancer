package loadbalancer

import (
	// "bufio"
	// "errors"
	// "fmt"
	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
	"log/slog"
	"net"
	// "strings"
)

type LoadBalancer struct {
	backend  []*backendserver.BackendServer
	logger   *slog.Logger
	listener net.Listener
}
