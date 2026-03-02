package loadbalancer

import (
	"sync/atomic"

	"github.com/anyangateny1/Load-Balancer/internal/backendserver"
)

type backend struct {
	server  *backendserver.BackendServer
	healthy atomic.Bool
}

func newBackend(server *backendserver.BackendServer) *backend {
	b := &backend{server: server}
	b.healthy.Store(true)
	return b
}
