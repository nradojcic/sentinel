package store

import (
	"sync"
	"time"
)

type NodeStats struct {
	CPU      float64
	RAM      float64
	LastSeen time.Time
}

type MonitorStore struct {
	mu    sync.RWMutex
	nodes map[string]NodeStats
}

func NewMonitorStore() *MonitorStore {
	return &MonitorStore{
		nodes: make(map[string]NodeStats),
	}
}

// Update adds or updates stats for a specific node
func (s *MonitorStore) Update(id string, cpu, ram float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes[id] = NodeStats{
		CPU:      cpu,
		RAM:      ram,
		LastSeen: time.Now(),
	}
}

// GetAll returns a snapshot of all node data
func (s *MonitorStore) GetAll() map[string]NodeStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := make(map[string]NodeStats)
	for k, v := range s.nodes {
		copy[k] = v
	}
	return copy
}
