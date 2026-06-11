package embedding

import (
	"crypto/sha256"
	"math"
	"strings"
	"unicode"
)

const Dimensions = 32

type HashEmbedder struct{}

func NewHashEmbedder() HashEmbedder {
	return HashEmbedder{}
}

// Embed creates a deterministic local vector for demonstrating pgvector.
func (HashEmbedder) Embed(text string) []float32 {
	vector := make([]float32, Dimensions)
	tokens := Tokens(text)
	for index, token := range tokens {
		addFeature(vector, token, 1)
		if index > 0 {
			addFeature(vector, tokens[index-1]+"_"+token, 0.5)
		}
	}

	normalize(vector)
	return vector
}

func Tokens(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(value rune) bool {
		return !unicode.IsLetter(value) && !unicode.IsDigit(value)
	})

	result := make([]string, 0, len(words))
	for _, word := range words {
		word = normalizeWord(word)
		if word == "" || stopWords[word] {
			continue
		}
		result = append(result, word)
	}
	return result
}

func normalizeWord(word string) string {
	switch word {
	case "favorite", "favourite", "favorites", "favourites",
		"like", "likes", "liked", "love", "loves", "prefer", "prefers", "preferred":
		return "preference"
	case "languages":
		return "language"
	case "databases":
		return "database"
	case "programming", "coding":
		return "code"
	}
	if len(word) > 4 && strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

func addFeature(vector []float32, feature string, weight float32) {
	hash := sha256.Sum256([]byte(feature))
	index := int(hash[0]) % Dimensions
	sign := float32(1)
	if hash[1]%2 == 1 {
		sign = -1
	}
	vector[index] += sign * weight
}

func normalize(vector []float32) {
	var magnitude float64
	for _, value := range vector {
		magnitude += float64(value * value)
	}
	if magnitude == 0 {
		return
	}
	scale := float32(math.Sqrt(magnitude))
	for index := range vector {
		vector[index] /= scale
	}
}

var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "at": true,
	"be": true, "do": true, "does": true, "for": true, "from": true,
	"i": true, "in": true, "is": true, "it": true, "me": true,
	"my": true, "of": true, "on": true, "or": true, "that": true,
	"the": true, "this": true, "to": true, "what": true, "which": true,
	"who": true, "why": true, "with": true, "you": true, "your": true,
}
