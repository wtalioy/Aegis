package stream

import "sync"

type Subscription[T any] struct {
	C      <-chan T
	Cancel func()
}

type Hub[T any] struct {
	mu   sync.RWMutex
	subs map[chan T]struct{}
}

func NewHub[T any]() *Hub[T] {
	return &Hub[T]{subs: make(map[chan T]struct{})}
}

func (h *Hub[T]) Subscribe(buffer int) Subscription[T] {
	if buffer <= 0 {
		buffer = 100
	}
	ch := make(chan T, buffer)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
		h.mu.Unlock()
	}

	return Subscription[T]{C: ch, Cancel: cancel}
}

func (h *Hub[T]) Publish(value T) {
	h.mu.RLock()
	subs := make([]chan T, 0, len(h.subs))
	for ch := range h.subs {
		subs = append(subs, ch)
	}
	h.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- value:
		default:
		}
	}
}
