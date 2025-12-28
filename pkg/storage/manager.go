package storage

import (
	"sync"
	"time"

	"aegis/pkg/events"
)

type Manager struct {
	store   *TimeRingBuffer
	indexer *Indexer
	mu      sync.RWMutex
}

func NewManager(capacity int, maxIndexSize int) *Manager {
	return &Manager{
		store:   NewTimeRingBuffer(capacity),
		indexer: NewIndexer(maxIndexSize),
	}
}

func (m *Manager) Append(event *Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.store.Append(event); err != nil {
		return err
	}

	m.indexer.IndexEvent(event)
	return nil
}

func (m *Manager) Query(start, end time.Time) ([]*Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.Query(start, end)
}

func (m *Manager) Latest(n int) ([]*Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.Latest(n)
}

func (m *Manager) QueryByPID(pid uint32) []*Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.indexer.QueryByPID(pid)
}

func (m *Manager) QueryByCgroup(cgroupID uint64) []*Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.indexer.QueryByCgroup(cgroupID)
}

func (m *Manager) QueryByType(eventType events.EventType) []*Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.indexer.QueryByType(eventType)
}

func (m *Manager) QueryByTypeWithTimeRange(eventType events.EventType, start, end time.Time, maxResults int) []*Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allEvents := m.indexer.QueryByType(eventType)
	if maxResults <= 0 {
		maxResults = 1000 // Default limit
	}

	filtered := make([]*Event, 0, len(allEvents))
	for _, ev := range allEvents {
		if ev != nil && !ev.Timestamp.Before(start) && !ev.Timestamp.After(end) {
			filtered = append(filtered, ev)
			if len(filtered) >= maxResults {
				break
			}
		}
	}
	return filtered
}

func (m *Manager) QueryByProcess(processName string) []*Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.indexer.QueryByProcess(processName)
}

func (m *Manager) QueryByFilter(filter Filter) []*Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.indexer.QueryByFilter(filter)
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store.Close()
}

func (m *Manager) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.Size()
}

func (m *Manager) Capacity() int {
	return m.store.Capacity()
}
