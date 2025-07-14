package backend

type LeastConnectionsBackend struct {
	*Backend
	activeConnections uint64
}

func (lcb *LeastConnectionsBackend) GetActiveConnections() uint64 {
	return lcb.activeConnections
}

func (lcb *LeastConnectionsBackend) IncConnections() {
	lcb.activeConnections++
}

func (lcb *LeastConnectionsBackend) DecConnections() {
	if lcb.activeConnections > 0 {
		lcb.activeConnections--
	}
}
