package serverpool

import (
	"log"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"loadbalancer/internal/backend"
	"loadbalancer/internal/utils"
)

type BackendPeer interface {
	NewBackend(url *url.URL, proxy *httputil.ReverseProxy) *backend.Backend
	SetIsAlive(alive bool)
	GetIsAlive() bool
	GetURL() *url.URL
	GetReverseProxy() *httputil.ReverseProxy
}

type RoundRobinList struct {
	backends []BackendPeer
	current  uint64
}

func (s *RoundRobinList) AddBackend(backend BackendPeer) {
	s.backends = append(s.backends, backend)
}

func (s *RoundRobinList) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.GetURL().String() == backendUrl.String() {
			b.SetIsAlive(alive)
			break
		}
	}
}

func (s *RoundRobinList) GetNextPeer() BackendPeer {
	next := s.nextIndex()
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

func (s *RoundRobinList) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := utils.IsBackendAlive(b.GetURL())
		b.SetIsAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.GetURL(), status)
	}
}

func (s *RoundRobinList) nextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}
