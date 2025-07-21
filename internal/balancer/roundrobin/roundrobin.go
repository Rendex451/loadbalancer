package roundrobin

import (
	"errors"
	"loadbalancer/internal/backend"
	"sync/atomic"
)

type RoundRobinBalancer struct {
	backends []*backend.Backend
	current  uint64
}

func New(backends []*backend.Backend) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: backends,
		current:  0,
	}
}

func (rrb *RoundRobinBalancer) nextIndex() int {
	if len(rrb.backends) == 0 {
		return -1
	}
	return int(atomic.AddUint64(&rrb.current, uint64(1)) % uint64(len(rrb.backends)))
}

func (rrb *RoundRobinBalancer) NextBackend() (*backend.Backend, error) {
	next := rrb.nextIndex()
	l := len(rrb.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(rrb.backends)
		b := rrb.backends[idx]
		if b.IsHealthy() {
			if i != next {
				atomic.StoreUint64(&rrb.current, uint64(idx))
			}
			b.IncConns()
			return b, nil
		}
	}
	return nil, errors.New("no healthy backends")
}

func (rrb *RoundRobinBalancer) ReleaseBackend(b *backend.Backend) {
	b.DecConns()
}

func (rrb *RoundRobinBalancer) UpdateBackendHealth(b *backend.Backend, isHealthy bool) {
	b.SetHealthy(isHealthy)
}
