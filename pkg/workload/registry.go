// Package workload provides a registry for tracking workloads (cgroups) and their activity.
package workload

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// WorkloadID is a unique identifier for a workload, backed by the kernel cgroup ID.
type WorkloadID uint64

// Metadata contains information about a workload.
type Metadata struct {
	ID           WorkloadID
	CgroupPath   string
	FirstSeen    time.Time
	LastSeen     time.Time
	ExecCount    int64
	FileCount    int64
	ConnectCount int64
	AlertCount   int64
}

// Registry tracks workloads and their activity.
type Registry struct {
	mu       sync.RWMutex
	data     map[WorkloadID]*Metadata
	lru      *list.List                   // LRU list for eviction
	lruIndex map[WorkloadID]*list.Element // Map from ID to LRU element
	maxSize  int
	count    atomic.Int32
}

// NewRegistry creates a new workload registry with the specified maximum size.
func NewRegistry(maxSize int) *Registry {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &Registry{
		data:     make(map[WorkloadID]*Metadata),
		lru:      list.New(),
		lruIndex: make(map[WorkloadID]*list.Element),
		maxSize:  maxSize,
	}
}

// RecordExec records an exec event for a workload.
func (r *Registry) RecordExec(cgroupID uint64, cgroupPath string) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.getOrCreate(id, cgroupPath)
	m.ExecCount++
	m.LastSeen = time.Now()
	r.touch(id)
}

// RecordFile records a file event for a workload.
func (r *Registry) RecordFile(cgroupID uint64, cgroupPath string) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.getOrCreate(id, cgroupPath)
	m.FileCount++
	m.LastSeen = time.Now()
	r.touch(id)
}

// RecordConnect records a connect event for a workload.
func (r *Registry) RecordConnect(cgroupID uint64, cgroupPath string) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.getOrCreate(id, cgroupPath)
	m.ConnectCount++
	m.LastSeen = time.Now()
	r.touch(id)
}

// RecordAlert records an alert for a workload.
func (r *Registry) RecordAlert(cgroupID uint64) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	if m, ok := r.data[id]; ok {
		m.AlertCount++
		m.LastSeen = time.Now()
		r.touch(id)
	}
}

// Get returns the metadata for a workload, or nil if not found.
func (r *Registry) Get(cgroupID uint64) *Metadata {
	id := WorkloadID(cgroupID)
	r.mu.RLock()
	defer r.mu.RUnlock()

	if m, ok := r.data[id]; ok {
		// Return a copy to avoid race conditions
		copy := *m
		return &copy
	}
	return nil
}

// List returns all workload metadata.
func (r *Registry) List() []Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Metadata, 0, len(r.data))
	for _, m := range r.data {
		result = append(result, *m)
	}
	return result
}

// Count returns the number of tracked workloads.
func (r *Registry) Count() int {
	return int(r.count.Load())
}

// getOrCreate returns the metadata for a workload, creating it if necessary.
// Must be called with the lock held.
func (r *Registry) getOrCreate(id WorkloadID, cgroupPath string) *Metadata {
	if m, ok := r.data[id]; ok {
		// Update path if we didn't have it before but now we do
		if m.CgroupPath == "" && cgroupPath != "" {
			m.CgroupPath = cgroupPath
		}
		return m
	}

	// Evict oldest if at capacity
	if len(r.data) >= r.maxSize {
		r.evictOldest()
	}

	now := time.Now()
	m := &Metadata{
		ID:         id,
		CgroupPath: cgroupPath,
		FirstSeen:  now,
		LastSeen:   now,
	}
	r.data[id] = m
	r.lruIndex[id] = r.lru.PushFront(id)
	r.count.Add(1)
	return m
}

// touch moves a workload to the front of the LRU list.
// Must be called with the lock held.
func (r *Registry) touch(id WorkloadID) {
	if elem, ok := r.lruIndex[id]; ok {
		r.lru.MoveToFront(elem)
	}
}

// evictOldest removes the least recently used workload.
// Must be called with the lock held.
func (r *Registry) evictOldest() {
	elem := r.lru.Back()
	if elem == nil {
		return
	}

	id := elem.Value.(WorkloadID)
	r.lru.Remove(elem)
	delete(r.lruIndex, id)
	delete(r.data, id)
	r.count.Add(-1)
}
