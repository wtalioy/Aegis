package chat

import (
	"sync"
	"time"

	"aegis/pkg/ai/types"
)

type Conversation struct {
	ID        string        `json:"id"`
	Messages  []types.Message `json:"messages"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type Store struct {
	mu      sync.RWMutex
	convos  map[string]*Conversation
	maxAge  time.Duration
	maxMsgs int
}

func NewStore() *Store {
	store := &Store{
		convos:  make(map[string]*Conversation),
		maxAge:  30 * time.Minute,
		maxMsgs: 20,
	}
	go store.cleanupLoop()
	return store
}

func (s *Store) GetOrCreate(sessionID string) *Conversation {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conv, ok := s.convos[sessionID]; ok {
		conv.UpdatedAt = time.Now()
		return conv
	}

	conv := &Conversation{
		ID:        sessionID,
		Messages:  make([]types.Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.convos[sessionID] = conv
	return conv
}

func (s *Store) AddMessage(sessionID string, msg types.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, ok := s.convos[sessionID]
	if !ok {
		return
	}

	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()
	if len(conv.Messages) > s.maxMsgs {
		conv.Messages = conv.Messages[len(conv.Messages)-s.maxMsgs:]
	}
}

func (s *Store) GetMessages(sessionID string) []types.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, ok := s.convos[sessionID]
	if !ok {
		return nil
	}

	result := make([]types.Message, len(conv.Messages))
	copy(result, conv.Messages)
	return result
}

func (s *Store) Clear(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.convos, sessionID)
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, conv := range s.convos {
			if now.Sub(conv.UpdatedAt) > s.maxAge {
				delete(s.convos, id)
			}
		}
		s.mu.Unlock()
	}
}

