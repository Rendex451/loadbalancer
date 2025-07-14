package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	url          *url.URL
	reverseProxy *httputil.ReverseProxy
	mux          sync.RWMutex
	isAlive      bool
}

type BackendPeer interface {
	NewBackend(url *url.URL, proxy *httputil.ReverseProxy) *Backend
	SetIsAlive(alive bool)
	GetIsAlive() bool
	GetURL() *url.URL
	GetReverseProxy() *httputil.ReverseProxy
}

func NewBackend(url *url.URL, proxy *httputil.ReverseProxy) *Backend {
	return &Backend{
		url:          url,
		reverseProxy: proxy,
		isAlive:      true,
	}
}

func (b *Backend) SetIsAlive(alive bool) {
	b.mux.Lock()
	b.isAlive = alive
	b.mux.Unlock()
}

func (b *Backend) GetIsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.isAlive
	b.mux.RUnlock()
	return
}

func (b *Backend) GetURL() *url.URL {
	return b.url
}

func (b *Backend) GetReverseProxy() *httputil.ReverseProxy {
	return b.reverseProxy
}
