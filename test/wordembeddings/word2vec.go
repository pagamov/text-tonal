package wordembeddings

import (
	"errors"
	"math"
)

// Word2Vec represents a simple word embedding model
type Word2Vec struct {
	embeddings map[string][]float64
}

// NewWord2Vec initializes a new Word2Vec model
func NewWord2Vec() *Word2Vec {
	return &Word2Vec{
		embeddings: make(map[string][]float64),
	}
}

// AddWord adds a word with its corresponding embedding
func (w *Word2Vec) AddWord(word string, embedding []float64) {
	w.embeddings[word] = embedding
}

// GetEmbedding retrieves the embedding for a given word
func (w *Word2Vec) GetEmbedding(word string) ([]float64, error) {
	embedding, exists := w.embeddings[word]
	if !exists {
		return nil, errors.New("word not found")
	}
	return embedding, nil
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, errors.New("embeddings must be of the same length")
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, errors.New("one of the vectors is zero")
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}
