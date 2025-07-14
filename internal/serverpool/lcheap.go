package serverpool

type LeastConnectionsBackend interface {
	BackendPeer
	GetActiveConnections() uint64
	IncConnections()
	DecConnections()
}

type LeastConnectionsHeap struct {
	backends []LeastConnectionsBackend
	current  uint64
}
