package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	memoryapp "github.com/example/memory-architecture-sample/internal/application/memory"
	"github.com/example/memory-architecture-sample/internal/domain"
	"github.com/example/memory-architecture-sample/internal/infrastructure/embedding"
	"github.com/example/memory-architecture-sample/internal/infrastructure/memory"
)

func TestDeleteRemovesLongTermMemory(t *testing.T) {
	store := memory.NewInMemoryLongTermStore()
	service := memoryapp.NewService(store, embedding.NewHashEmbedder())
	item := domain.SemanticMemory{
		ID:             "memory-1",
		ConversationID: "conversation-1",
		Content:        "My favorite language is Go.",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	if err := store.Save(context.Background(), item, embedding.NewHashEmbedder().Embed(item.Content)); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}

	if err := service.Delete(context.Background(), item.ID); err != nil {
		t.Fatalf("Delete returned an error: %v", err)
	}
	if err := service.Delete(context.Background(), item.ID); !errors.Is(err, memoryapp.ErrMemoryNotFound) {
		t.Fatalf("expected ErrMemoryNotFound, got %v", err)
	}
}

func TestClearConversationKeepsOtherConversationMemories(t *testing.T) {
	store := memory.NewInMemoryLongTermStore()
	embedder := embedding.NewHashEmbedder()
	service := memoryapp.NewService(store, embedder)
	for _, item := range []domain.SemanticMemory{
		{ID: "memory-1", ConversationID: "conversation-1", Content: "Go", ExpiresAt: time.Now().Add(time.Hour)},
		{ID: "memory-2", ConversationID: "conversation-2", Content: "PostgreSQL", ExpiresAt: time.Now().Add(time.Hour)},
	} {
		if err := store.Save(context.Background(), item, embedder.Embed(item.Content)); err != nil {
			t.Fatalf("Save returned an error: %v", err)
		}
	}

	if err := service.ClearConversation(context.Background(), "conversation-1"); err != nil {
		t.Fatalf("ClearConversation returned an error: %v", err)
	}
	first, _ := service.Search(context.Background(), "conversation-1", "Go", 5)
	second, _ := service.Search(context.Background(), "conversation-2", "PostgreSQL", 5)
	if len(first) != 0 {
		t.Fatalf("expected conversation-1 to be empty, got %d memories", len(first))
	}
	if len(second) != 1 {
		t.Fatalf("expected conversation-2 memory to remain, got %d", len(second))
	}
}

func TestSearchRanksRelatedPreferenceFirst(t *testing.T) {
	store := memory.NewInMemoryLongTermStore()
	embedder := embedding.NewHashEmbedder()
	service := memoryapp.NewService(store, embedder)
	items := []domain.SemanticMemory{
		{ID: "language", ConversationID: "demo", Content: "My favorite programming language is Go.", ExpiresAt: time.Now().Add(time.Hour)},
		{ID: "database", ConversationID: "demo", Content: "I use PostgreSQL as my database.", ExpiresAt: time.Now().Add(time.Hour)},
		{ID: "city", ConversationID: "demo", Content: "I live in Tehran.", ExpiresAt: time.Now().Add(time.Hour)},
	}
	for _, item := range items {
		if err := store.Save(context.Background(), item, embedder.Embed(item.Content)); err != nil {
			t.Fatalf("Save returned an error: %v", err)
		}
	}

	result, err := service.Search(
		context.Background(),
		"demo",
		"Which programming language do I like?",
		3,
	)
	if err != nil {
		t.Fatalf("Search returned an error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected a matching memory")
	}
	if result[0].ID != "language" {
		t.Fatalf("expected language memory first, got %q", result[0].ID)
	}
}

func TestSearchDoesNotReturnUnrelatedWeakMatches(t *testing.T) {
	store := memory.NewInMemoryLongTermStore()
	embedder := embedding.NewHashEmbedder()
	service := memoryapp.NewService(store, embedder)
	item := domain.SemanticMemory{
		ID:             "city",
		ConversationID: "demo",
		Content:        "I live in Tehran.",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	if err := store.Save(context.Background(), item, embedder.Embed(item.Content)); err != nil {
		t.Fatalf("Save returned an error: %v", err)
	}

	result, err := service.Search(context.Background(), "demo", "favorite database", 5)
	if err != nil {
		t.Fatalf("Search returned an error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected no relevant memories, got %d", len(result))
	}
}
