package serverpool

import (
	"log"
	"net/url"
	"sync/atomic"

	"loadbalancer/internal/backend"
	"loadbalancer/internal/utils"
)

type ServerPool struct {
	backends []*backend.Backend
	current  uint64
}

func (s *ServerPool) AddBackend(backend *backend.Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetIsAlive(alive)
			break
		}
	}
}

func (s *ServerPool) GetNextPeer() *backend.Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].GetIsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := utils.IsBackendAlive(b.URL)
		b.SetIsAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}
