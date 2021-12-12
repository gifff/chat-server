package websocket

import (
	"sync"
)

// ConnectionPool holds pool of ConnectionDispatcher with concurrent-safe handling for fast storing and deletion
type ConnectionPool struct {
	lastID int
	mu     sync.Mutex
	pool   map[int]ConnectionDispatcher
}

// NewConnectionPool returns ConnectionPool with prepared internal map
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pool: make(map[int]ConnectionDispatcher),
	}
}

// Store assigns the ConnectionDispatcher into internal pool with generated ID and return the ID for deletion
func (cp *ConnectionPool) Store(conn ConnectionDispatcher) int {
	cp.mu.Lock()
	id := cp.lastID
	cp.pool[id] = conn
	cp.lastID++
	cp.mu.Unlock()
	return id
}

// Get returns ConnectionDispatcher from pool by ID
func (cp *ConnectionPool) Get(id int) ConnectionDispatcher {
	cp.mu.Lock()
	conn, ok := cp.pool[id]
	cp.mu.Unlock()
	if !ok {
		return nil
	}
	return conn
}

// Delete deletes ConnectionDispatcher from pool based on given pool map ID
func (cp *ConnectionPool) Delete(id int) {
	cp.mu.Lock()
	delete(cp.pool, id)
	cp.mu.Unlock()
}

// Slice flattens the internal pool map into slice of connections
func (cp *ConnectionPool) Slice() []ConnectionDispatcher {
	cp.mu.Lock()
	conns := make([]ConnectionDispatcher, len(cp.pool))
	i := 0
	for k := range cp.pool {
		conns[i] = cp.pool[k]
		i++
	}
	cp.mu.Unlock()
	return conns
}
