package loadbalancer

import (
	"log/slog"
	"net"
	"time"
)

type healthChecker struct {
	backends []*backend
	interval time.Duration
	timeout  time.Duration
	logger   *slog.Logger
	done     chan struct{}
}

func newHealthChecker(backends []*backend, interval time.Duration, logger *slog.Logger) *healthChecker {
	return &healthChecker{
		backends: backends,
		interval: interval,
		timeout:  2 * time.Second,
		logger:   logger,
		done:     make(chan struct{}),
	}
}

func (hc *healthChecker) probe(addr net.Addr) bool {
	conn, err := net.DialTimeout("tcp", addr.String(), hc.timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (hc *healthChecker) run() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, b := range hc.backends {
				wasHealthy := b.healthy.Load()
				isHealthy := hc.probe(b.server.Addr())
				b.healthy.Store(isHealthy)

				if wasHealthy && !isHealthy {
					hc.logger.Warn("backend went down", "addr", b.server.Addr())
				} else if !wasHealthy && isHealthy {
					hc.logger.Info("backend recovered", "addr", b.server.Addr())
				}
			}
		case <-hc.done:
			return
		}
	}
}

func (hc *healthChecker) stop() {
	close(hc.done)
}
