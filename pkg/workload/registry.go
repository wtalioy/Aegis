package workload

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type WorkloadID uint64

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

type Registry struct {
	mu       sync.RWMutex
	data     map[WorkloadID]*Metadata
	lru      *list.List
	lruIndex map[WorkloadID]*list.Element
	maxSize  int
	count    atomic.Int32
}

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

func (r *Registry) RecordExec(cgroupID uint64, cgroupPath string) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.getOrCreate(id, cgroupPath)
	m.ExecCount++
	m.LastSeen = time.Now()
	r.touch(id)
}

func (r *Registry) RecordFile(cgroupID uint64, cgroupPath string) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.getOrCreate(id, cgroupPath)
	m.FileCount++
	m.LastSeen = time.Now()
	r.touch(id)
}

func (r *Registry) RecordConnect(cgroupID uint64, cgroupPath string) {
	id := WorkloadID(cgroupID)
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.getOrCreate(id, cgroupPath)
	m.ConnectCount++
	m.LastSeen = time.Now()
	r.touch(id)
}

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

func (r *Registry) Get(cgroupID uint64) *Metadata {
	id := WorkloadID(cgroupID)
	r.mu.RLock()
	defer r.mu.RUnlock()

	if m, ok := r.data[id]; ok {
		copy := *m
		return &copy
	}
	return nil
}

func (r *Registry) List() []Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Metadata, 0, len(r.data))
	for _, m := range r.data {
		result = append(result, *m)
	}
	return result
}

func (r *Registry) Count() int {
	return int(r.count.Load())
}

func (r *Registry) getOrCreate(id WorkloadID, cgroupPath string) *Metadata {
	if m, ok := r.data[id]; ok {
		if m.CgroupPath == "" && cgroupPath != "" {
			m.CgroupPath = cgroupPath
		}
		return m
	}

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

func (r *Registry) touch(id WorkloadID) {
	if elem, ok := r.lruIndex[id]; ok {
		r.lru.MoveToFront(elem)
	}
}

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
