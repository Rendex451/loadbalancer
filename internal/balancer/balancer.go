package balancer

import "loadbalancer/internal/backend"

type Balancer interface {
	NextBackend() (*backend.Backend, error)
	ReleaseBackend(backend *backend.Backend)
	UpdateBackendHealth(backend *backend.Backend, isHealthy bool)
}
