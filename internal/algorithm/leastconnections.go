package algorithm

import (
	"math/rand"
	"sync"
)

type Random struct {
	mu  sync.Mutex // protects rng
	rng *rand.Rand
}

func NewRandom() *Random {
	return &Random{
		rng: rand.New(rand.NewSource(rand.Int63())),
	}
}

func (r *Random) Next(numBackends int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rng.Intn(numBackends)
}
