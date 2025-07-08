package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex
	IsAlive      bool
}

func (b *Backend) SetIsAlive(alive bool) {
	b.mux.Lock()
	b.IsAlive = alive
	b.mux.Unlock()
}

func (b *Backend) GetIsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.IsAlive
	b.mux.RUnlock()
	return
}
