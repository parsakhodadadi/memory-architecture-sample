package ports

import (
	"context"

	"github.com/example/memory-architecture-sample/internal/domain"
)

type ShortTermMemory interface {
	Save(ctx context.Context, messages ...domain.Message) error
	List(ctx context.Context, conversationID string, limit int) ([]domain.Message, error)
	Clear(ctx context.Context, conversationID string) error
	Ping(ctx context.Context) error
}

type LongTermMemory interface {
	Save(ctx context.Context, memory domain.SemanticMemory, embedding []float32) error
	Search(ctx context.Context, conversationID, query string, embedding []float32, limit int) ([]domain.SemanticMemory, error)
	Delete(ctx context.Context, memoryID string) (bool, error)
	ClearConversation(ctx context.Context, conversationID string) error
	Ping(ctx context.Context) error
}

type Embedder interface {
	Embed(text string) []float32
}

type Responder interface {
	Reply(
		ctx context.Context,
		userMessage string,
		recent []domain.Message,
		recalled []domain.SemanticMemory,
	) (string, error)
}
