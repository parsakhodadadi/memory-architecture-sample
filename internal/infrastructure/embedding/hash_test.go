package embedding_test

import (
	"testing"

	"github.com/example/memory-architecture-sample/internal/infrastructure/embedding"
)

func TestHashEmbedderIsDeterministic(t *testing.T) {
	embedder := embedding.NewHashEmbedder()
	first := embedder.Embed("Memory architecture with Go")
	second := embedder.Embed("Memory architecture with Go")

	if len(first) != embedding.Dimensions {
		t.Fatalf("expected %d dimensions, got %d", embedding.Dimensions, len(first))
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("embedding differs at index %d", index)
		}
	}
}

func TestTokensNormalizeQuestionsAndPreferenceWords(t *testing.T) {
	tokens := embedding.Tokens("Which programming language do I like?")
	expected := []string{"code", "language", "preference"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, tokens)
	}
	for index := range expected {
		if tokens[index] != expected[index] {
			t.Fatalf("expected %v, got %v", expected, tokens)
		}
	}
}
