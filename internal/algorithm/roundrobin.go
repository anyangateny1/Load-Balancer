package algorithm

import (
	"sync/atomic"
)

type RoundRobin struct {
	counter atomic.Uint32
}

func (rr *RoundRobin) Next(numBackends int) int {
	index := rr.counter.Add(1) - 1
	return int(index) % numBackends
}
