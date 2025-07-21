package main

import (
	"flag"
	"loadbalancer/internal/backend"
	"loadbalancer/internal/balancer"
	"loadbalancer/internal/balancer/leastconnections"
	"loadbalancer/internal/balancer/roundrobin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func main() {
	var servers string
	var port int
	var lbType string

	flag.StringVar(&servers, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 3030, "Port to serve")
	flag.StringVar(&lbType, "lbtype", "rr", "Load balancer type: rr (round robin) or lc (least connections)")
	flag.Parse()

	serverList := strings.Split(servers, ",")
	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	var backends []*backend.Backend
	for _, s := range serverList {
		u, err := url.Parse(s)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			w.WriteHeader(http.StatusBadGateway)
		}

		backends = append(backends, backend.New(u, proxy))
	}

	// 2. Создание балансировщика
	var lb balancer.Balancer
	switch lbType {
	case "rr":
		lb = roundrobin.New(backends)
	case "lc":
		lb = leastconnections.New(backends)
	default:
		log.Fatalf("Unknown load balancer type: %s", lbType)
	}

	// 3. Запуск health check в фоне
	go healthCheckRoutine(lb, backends, 10*time.Second)

	// 4. HTTP сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		backend, err := lb.NextBackend()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer lb.ReleaseBackend(backend)

		backend.ReverseProxy.ServeHTTP(w, r)
	})

	log.Printf("Load balancer started on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func healthCheckRoutine(lb balancer.Balancer, backends []*backend.Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		for _, b := range backends {
			wasHealthy := b.IsHealthy()
			isHealthy := b.HealthCheck()

			if wasHealthy != isHealthy {
				lb.UpdateBackendHealth(b, isHealthy)
				status := "DOWN"
				if isHealthy {
					status = "UP"
				}
				log.Printf("Backend %s status changed: %s", b.URL.String(), status)
			}

			log.Printf("Backend %s [conns: %d] [healthy: %v]", b.URL.String(), b.GetConns(), b.IsHealthy())
		}
	}
}
