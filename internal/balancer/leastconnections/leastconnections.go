package leastconnections

import (
	"container/heap"
	"errors"
	"loadbalancer/internal/backend"
	"sync"
)

type backendHeap struct {
	backends []*backend.Backend
	indexes  map[*backend.Backend]int
}

func (h *backendHeap) Len() int { return len(h.backends) }
func (h *backendHeap) Less(i, j int) bool {
	if h.backends[i].IsHealthy() != h.backends[j].IsHealthy() {
		return h.backends[i].IsHealthy()
	}
	return h.backends[i].GetConns() < h.backends[j].GetConns()
}
func (h *backendHeap) Swap(i, j int) {
	h.backends[i], h.backends[j] = h.backends[j], h.backends[i]
	h.indexes[h.backends[i]] = i
	h.indexes[h.backends[j]] = j
}
func (h *backendHeap) Push(x any) {
	b := x.(*backend.Backend)
	h.indexes[b] = len(h.backends)
	h.backends = append(h.backends, b)
}
func (h *backendHeap) Pop() any {
	n := len(h.backends)
	b := h.backends[n-1]
	h.backends = h.backends[:n-1]
	delete(h.indexes, b)
	return b
}

type LeastConnectionsBalancer struct {
	heap *backendHeap
	mux  sync.RWMutex
}

func New(backends []*backend.Backend) *LeastConnectionsBalancer {
	h := &backendHeap{
		backends: make([]*backend.Backend, 0, len(backends)),
		indexes:  make(map[*backend.Backend]int),
	}

	for _, b := range backends {
		heap.Push(h, b)
	}

	return &LeastConnectionsBalancer{
		heap: h,
	}
}

func (lcb *LeastConnectionsBalancer) NextBackend() (*backend.Backend, error) {
	lcb.mux.Lock()
	defer lcb.mux.Unlock()

	if lcb.heap.Len() == 0 {
		return nil, errors.New("no backends available")
	}

	b := lcb.heap.backends[0]
	if !b.IsHealthy() {
		return nil, errors.New("no healthy backends")
	}

	b.IncConns()
	heap.Fix(lcb.heap, 0)
	return b, nil
}

func (lcb *LeastConnectionsBalancer) ReleaseBackend(b *backend.Backend) {
	b.DecConns()

	lcb.mux.Lock()
	defer lcb.mux.Unlock()

	if index, exists := lcb.heap.indexes[b]; exists {
		heap.Fix(lcb.heap, index)
	}
}

func (lcb *LeastConnectionsBalancer) UpdateBackendHealth(b *backend.Backend, isHealthy bool) {
	lcb.mux.Lock()
	defer lcb.mux.Unlock()

	if index, exists := lcb.heap.indexes[b]; exists {
		b.SetHealthy(isHealthy)
		heap.Fix(lcb.heap, index)
	}
}
