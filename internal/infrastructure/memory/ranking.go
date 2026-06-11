package memory

import (
	"sort"

	"github.com/example/memory-architecture-sample/internal/domain"
	"github.com/example/memory-architecture-sample/internal/infrastructure/embedding"
)

const minimumSearchScore = 0.20

func rankMemories(query string, memories []domain.SemanticMemory, limit int) []domain.SemanticMemory {
	queryTokens := tokenSet(embedding.Tokens(query))
	ranked := make([]domain.SemanticMemory, 0, len(memories))

	for _, memory := range memories {
		lexical := queryCoverage(queryTokens, tokenSet(embedding.Tokens(memory.Content)))
		vectorScore := (memory.Similarity + 1) / 2
		if vectorScore < 0 {
			vectorScore = 0
		}
		if vectorScore > 1 {
			vectorScore = 1
		}

		memory.Similarity = 0.70*lexical + 0.30*vectorScore
		if memory.Similarity >= minimumSearchScore {
			ranked = append(ranked, memory)
		}
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].Similarity > ranked[j].Similarity
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked
}

func tokenSet(tokens []string) map[string]struct{} {
	result := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		result[token] = struct{}{}
	}
	return result
}

func queryCoverage(query, content map[string]struct{}) float64 {
	if len(query) == 0 {
		return 0
	}
	var matches int
	for token := range query {
		if _, exists := content[token]; exists {
			matches++
		}
	}
	return float64(matches) / float64(len(query))
}
