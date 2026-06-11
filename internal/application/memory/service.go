package memory

import (
	"context"
	"errors"
	"strings"

	"github.com/example/memory-architecture-sample/internal/domain"
	"github.com/example/memory-architecture-sample/internal/ports"
)

var (
	ErrConversationIDRequired = errors.New("conversationId is required")
	ErrMemoryIDRequired       = errors.New("memoryId is required")
	ErrMemoryNotFound         = errors.New("memory not found")
	ErrQueryRequired          = errors.New("query is required")
)

type Service struct {
	longTerm ports.LongTermMemory
	embedder ports.Embedder
}

func NewService(longTerm ports.LongTermMemory, embedder ports.Embedder) *Service {
	return &Service{longTerm: longTerm, embedder: embedder}
}

func (s *Service) Search(
	ctx context.Context,
	conversationID string,
	query string,
	limit int,
) ([]domain.SemanticMemory, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return nil, ErrConversationIDRequired
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrQueryRequired
	}
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	return s.longTerm.Search(ctx, conversationID, query, s.embedder.Embed(query), limit)
}

func (s *Service) Delete(ctx context.Context, memoryID string) error {
	memoryID = strings.TrimSpace(memoryID)
	if memoryID == "" {
		return ErrMemoryIDRequired
	}
	deleted, err := s.longTerm.Delete(ctx, memoryID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrMemoryNotFound
	}
	return nil
}

func (s *Service) ClearConversation(ctx context.Context, conversationID string) error {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return ErrConversationIDRequired
	}
	return s.longTerm.ClearConversation(ctx, conversationID)
}
