package algorithm

import (
	"sync/atomic"
)

type RoundRobin struct {
	counter uint32
}

func (rr *RoundRobin) Next(numBackends int) int {
	index := atomic.AddUint32(&rr.counter, 1) - 1
	return int(index) % numBackends
}
