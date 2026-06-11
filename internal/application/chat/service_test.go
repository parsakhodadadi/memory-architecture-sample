package chat_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/memory-architecture-sample/internal/application/chat"
	"github.com/example/memory-architecture-sample/internal/infrastructure/embedding"
	"github.com/example/memory-architecture-sample/internal/infrastructure/memory"
	"github.com/example/memory-architecture-sample/internal/infrastructure/responder"
)

func TestSendStoresUserAndAssistantMessages(t *testing.T) {
	shortTerm := memory.NewInMemoryStore()
	longTerm := memory.NewInMemoryLongTermStore()
	now := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	service := chat.NewService(
		shortTerm,
		longTerm,
		embedding.NewHashEmbedder(),
		responder.NewTemplateResponder(),
		func() time.Time { return now },
		30*24*time.Hour,
	)

	output, err := service.Send(context.Background(), chat.SendInput{
		ConversationID: "conversation-1",
		Message:        "I prefer Go.",
	})
	if err != nil {
		t.Fatalf("Send returned an error: %v", err)
	}
	if output.Reply == "" {
		t.Fatal("expected a reply")
	}

	history, err := service.History(context.Background(), "conversation-1", 20)
	if err != nil {
		t.Fatalf("History returned an error: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(history))
	}
	if history[0].Content != "I prefer Go." {
		t.Fatalf("unexpected user message: %q", history[0].Content)
	}
}

func TestSendRejectsEmptyMessage(t *testing.T) {
	service := chat.NewService(
		memory.NewInMemoryStore(),
		memory.NewInMemoryLongTermStore(),
		embedding.NewHashEmbedder(),
		responder.NewTemplateResponder(),
		time.Now,
		30*24*time.Hour,
	)

	_, err := service.Send(context.Background(), chat.SendInput{
		ConversationID: "conversation-1",
		Message:        " ",
	})
	if err != chat.ErrMessageRequired {
		t.Fatalf("expected ErrMessageRequired, got %v", err)
	}
}

func TestSendRecallsRelatedLongTermMemory(t *testing.T) {
	shortTerm := memory.NewInMemoryStore()
	longTerm := memory.NewInMemoryLongTermStore()
	service := chat.NewService(
		shortTerm,
		longTerm,
		embedding.NewHashEmbedder(),
		responder.NewTemplateResponder(),
		time.Now,
		30*24*time.Hour,
	)

	_, err := service.Send(context.Background(), chat.SendInput{
		ConversationID: "conversation-1",
		Message:        "My favorite language is Go.",
	})
	if err != nil {
		t.Fatalf("first Send returned an error: %v", err)
	}

	output, err := service.Send(context.Background(), chat.SendInput{
		ConversationID: "conversation-1",
		Message:        "What is my favorite language?",
	})
	if err != nil {
		t.Fatalf("second Send returned an error: %v", err)
	}
	if len(output.RecalledMemory) == 0 {
		t.Fatal("expected a recalled long-term memory")
	}
	if output.RecalledMemory[0].Content != "My favorite language is Go." {
		t.Fatalf("unexpected recalled memory: %q", output.RecalledMemory[0].Content)
	}
}
