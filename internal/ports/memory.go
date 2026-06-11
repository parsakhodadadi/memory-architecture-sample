package ports

import (
	"context"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type ShortTermMemory interface {
	Save(ctx context.Context, messages ...domain.Message) error
	List(ctx context.Context, conversationID string, limit int) ([]domain.Message, error)
	Clear(ctx context.Context, conversationID string) error
}

type Responder interface {
	Reply(ctx context.Context, userMessage string, recent []domain.Message) (string, error)
}
