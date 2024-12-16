package main

import (
	"fmt"
	"test/wordembeddings"
)

func main() {
	// Create a new Word2Vec model
	model := wordembeddings.NewWord2Vec()

	// Add some words and their embeddings
	model.AddWord("king", []float64{0.5, 0.1, 0.3})
	model.AddWord("queen", []float64{0.4, 0.2, 0.5})
	model.AddWord("man", []float64{0.6, 0.1, 0.2})
	model.AddWord("woman", []float64{0.3, 0.4, 0.5})

	// Retrieve an embedding
	embedding, err := model.GetEmbedding("king")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Embedding for 'king':", embedding)
	}

	// Calculate cosine similarity
	sim, err := wordembeddings.CosineSimilarity(
		[]float64{0.5, 0.1, 0.3},
		[]float64{0.4, 0.2, 0.5},
	)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Cosine similarity between 'king' and 'queen': %.4f\n", sim)
	}
}
