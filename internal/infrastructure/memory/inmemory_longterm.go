package memory

import (
	"context"
	"math"
	"sync"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type inMemoryVector struct {
	memory    domain.SemanticMemory
	embedding []float32
}

type InMemoryLongTermStore struct {
	mu      sync.RWMutex
	entries []inMemoryVector
}

func NewInMemoryLongTermStore() *InMemoryLongTermStore {
	return &InMemoryLongTermStore{}
}

func (s *InMemoryLongTermStore) Save(
	_ context.Context,
	memory domain.SemanticMemory,
	embedding []float32,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	copyOfEmbedding := append([]float32(nil), embedding...)
	s.entries = append(s.entries, inMemoryVector{memory: memory, embedding: copyOfEmbedding})
	return nil
}

func (s *InMemoryLongTermStore) Search(
	_ context.Context,
	conversationID string,
	query string,
	embedding []float32,
	limit int,
) ([]domain.SemanticMemory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.SemanticMemory, 0)
	for _, entry := range s.entries {
		if entry.memory.ConversationID != conversationID {
			continue
		}
		memory := entry.memory
		memory.Similarity = cosineSimilarity(entry.embedding, embedding)
		result = append(result, memory)
	}
	return rankMemories(query, result, limit), nil
}

func (s *InMemoryLongTermStore) Delete(_ context.Context, memoryID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, entry := range s.entries {
		if entry.memory.ID != memoryID {
			continue
		}
		s.entries = append(s.entries[:index], s.entries[index+1:]...)
		return true, nil
	}
	return false, nil
}

func (s *InMemoryLongTermStore) ClearConversation(_ context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	remaining := s.entries[:0]
	for _, entry := range s.entries {
		if entry.memory.ConversationID != conversationID {
			remaining = append(remaining, entry)
		}
	}
	s.entries = remaining
	return nil
}

func (s *InMemoryLongTermStore) Ping(_ context.Context) error {
	return nil
}

func cosineSimilarity(left, right []float32) float64 {
	var dot, leftMagnitude, rightMagnitude float64
	for index := 0; index < len(left) && index < len(right); index++ {
		dot += float64(left[index] * right[index])
		leftMagnitude += float64(left[index] * left[index])
		rightMagnitude += float64(right[index] * right[index])
	}
	if leftMagnitude == 0 || rightMagnitude == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftMagnitude) * math.Sqrt(rightMagnitude))
}
