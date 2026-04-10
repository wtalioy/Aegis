package insights

import (
	"sort"
	"sync"
	"time"
)

type Subscription[T any] struct {
	C      <-chan T
	Cancel func()
}

// Store owns Insight persistence, retention/dedup policy, sorting, and pub-sub.
// It is generic so it does not need to import the sentinel package (avoids cycles).
type Store[T any] struct {
	mu          sync.RWMutex
	insights    []T
	subscribers map[chan T]struct{}
}

func NewStore[T any]() *Store[T] {
	return &Store[T]{
		insights:    make([]T, 0),
		subscribers: make(map[chan T]struct{}),
	}
}

func (s *Store[T]) Reset() {
	s.mu.Lock()
	s.insights = make([]T, 0)
	s.mu.Unlock()
}

func (s *Store[T]) Add(insights []T, less func(a, b T) bool) {
	if len(insights) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Append
	s.insights = append(s.insights, insights...)

	// Sort newest first (caller provides comparator)
	if less != nil {
		sort.Slice(s.insights, func(i, j int) bool { return less(s.insights[i], s.insights[j]) })
	}

	// Keep last 1000 (newest first, so trim tail).
	if len(s.insights) > 1000 {
		s.insights = s.insights[:1000]
	}

	// Snapshot subscribers.
	subs := make([]chan T, 0, len(s.subscribers))
	for ch := range s.subscribers {
		subs = append(subs, ch)
	}

	// Fan-out: best effort, non-blocking.
	for _, in := range insights {
		for _, ch := range subs {
			select {
			case ch <- in:
			default:
			}
		}
	}
}

func (s *Store[T]) List(limit int) []T {
	s.mu.RLock()
	out := make([]T, len(s.insights))
	copy(out, s.insights)
	s.mu.RUnlock()

	if limit <= 0 || limit > len(out) {
		return out
	}
	return out[:limit]
}

func (s *Store[T]) Subscribe(buffer int) Subscription[T] {
	if buffer <= 0 {
		buffer = 100
	}
	ch := make(chan T, buffer)

	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		if _, ok := s.subscribers[ch]; ok {
			delete(s.subscribers, ch)
			close(ch)
		}
		s.mu.Unlock()
	}

	return Subscription[T]{C: ch, Cancel: cancel}
}

// KeepAlive is a small helper to avoid unused import of time in some build tags.
// (time is used in other files in this package).
var _ = time.Now
