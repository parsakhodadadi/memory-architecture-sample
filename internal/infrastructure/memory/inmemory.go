package memory

import (
	"context"
	"sync"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type InMemoryStore struct {
	mu       sync.RWMutex
	messages map[string][]domain.Message
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{messages: make(map[string][]domain.Message)}
}

func (s *InMemoryStore) Save(_ context.Context, messages ...domain.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, message := range messages {
		s.messages[message.ConversationID] = append(s.messages[message.ConversationID], message)
	}
	return nil
}

func (s *InMemoryStore) List(_ context.Context, conversationID string, limit int) ([]domain.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := s.messages[conversationID]
	start := 0
	if limit > 0 && len(messages) > limit {
		start = len(messages) - limit
	}

	result := make([]domain.Message, len(messages)-start)
	copy(result, messages[start:])
	return result, nil
}

func (s *InMemoryStore) Clear(_ context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.messages, conversationID)
	return nil
}

func (s *InMemoryStore) Ping(_ context.Context) error {
	return nil
}
