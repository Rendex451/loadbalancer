package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	connsCount   atomic.Int32
	healthy      atomic.Bool
	lastCheck    atomic.Int64 // unix timestamp
}

func New(url *url.URL, proxy *httputil.ReverseProxy) *Backend {
	b := &Backend{
		URL:          url,
		ReverseProxy: proxy,
	}
	b.healthy.Store(true)
	return b
}

func (b *Backend) IncConns() {
	b.connsCount.Add(1)
}

func (b *Backend) DecConns() {
	if b.connsCount.Load() > 0 {
		b.connsCount.Add(-1)
	}
}

func (b *Backend) GetConns() int32 {
	return b.connsCount.Load()
}

func (b *Backend) SetHealthy(healthy bool) {
	b.healthy.Store(healthy)
}

func (b *Backend) IsHealthy() bool {
	return b.healthy.Load()
}

func (b *Backend) HealthCheck() bool {
	// Кеширование на 5 секунд
	if time.Now().Unix()-b.lastCheck.Load() < 5 {
		return b.healthy.Load()
	}

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(b.URL.String() + "/health")

	healthy := err == nil && resp.StatusCode == http.StatusOK
	b.healthy.Store(healthy)
	b.lastCheck.Store(time.Now().Unix())

	return healthy
}
