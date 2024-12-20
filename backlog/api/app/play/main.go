package main

import (
	"fmt"
	"log"
	"strings"

	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

// Data structure to hold the training data and labels
type Data struct {
	Texts  []string
	Labels []int
}

// Simple function to preprocess text (tokenization and vectorization)
func preprocess(texts []string) (tensor.Tensor, map[string]int) {
	// Create a vocabulary map
	vocab := make(map[string]int)
	for _, text := range texts {
		words := strings.Fields(text)
		for _, word := range words {
			if _, exists := vocab[word]; !exists {
				vocab[word] = len(vocab)
			}
		}
	}

	// Create a tensor to hold the vectorized data
	data := tensor.New(tensor.WithShape(len(texts), len(vocab)), tensor.WithBacking(make([]float32, len(texts)*len(vocab))))

	for i, text := range texts {
		words := strings.Fields(text)
		for _, word := range words {
			if index, exists := vocab[word]; exists {
				data.SetAt(1.0, i, index) // One-hot encoding
			}
		}
	}

	return data, vocab
}

// Build a simple feedforward neural network
func buildModel(g *gorgonia.ExprGraph, input tensor.Tensor, numClasses int) (*gorgonia.Node, *gorgonia.Node, *gorgonia.Node, error) {
	x := gorgonia.NewMatrix(g, input.Dtype(), gorgonia.WithShape(input.Shape()...), gorgonia.WithName("x"))
	w := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(input.Shape()[1], numClasses), gorgonia.WithName("w"), gorgonia.WithLearnable(true))
	b := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(numClasses), gorgonia.WithName("b"), gorgonia.WithLearnable(true))

	// Forward pass
	logits := gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(x, w)), b))
	prob := gorgonia.Must(gorgonia.SoftMax(logits))

	return prob, w, b, nil
}

// Train the model
func trainModel(g *gorgonia.ExprGraph, prob *gorgonia.Node, labels tensor.Tensor, w *gorgonia.Node, b *gorgonia.Node) {
	// Define the loss function
	y := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(labels.Shape()...), gorgonia.WithName("y"))
	loss := gorgonia.Must(gorgonia.Neg(gorgonia.Must(gorgonia.Sum(gorgonia.Must(gorgonia.Mul(y, gorgonia.Must(gorgonia.Log(prob))))))))

	// Create a VM and run the training
	vm := gorgonia.NewTapeMachine(g)

	// Set the labels for the training
	gorgonia.Read(labels, y)

	// Run the forward pass and compute the loss
	if err := vm.RunAll(); err != nil {
		log.Fatal(err)
	}

	// Backpropagation
	if err := gorgonia.Backward(loss, w, b); err != nil {
		log.Fatal(err)
	}

	// Update weights (this is a simple example, you would typically use an optimizer)
	// Here we are not actually updating weights, just demonstrating the flow
}

func main() {
	// Sample training data
	data := Data{
		Texts: []string{
			"the cat sat on the mat",
			"the dog sat on the log",
			"the cat and the dog are friends",
			"dogs are great pets",
			"cats are also great pets",
		},
		Labels: []int{0, 1, 0, 1, 0}, // 0 for cat, 1 for dog
	}

	// Preprocess the data
	inputTensor, vocab := preprocess(data.Texts)

	// Create a new computation graph
	g := gorgonia.NewGraph()

	// Build the model
	prob, w, b, err := buildModel(g, inputTensor, 2) // 2 classes: cat and dog
	if err != nil {
		log.Fatal(err)
	}

	// Create a tensor for labels (one-hot encoding ```go
	labelTensor := tensor.New(tensor.WithShape(len(data.Labels), 2), tensor.WithBacking(make([]float32, len(data.Labels)*2)))
	for i, label := range data.Labels {
		labelTensor.SetAt(1.0, i, label) // One-hot encoding
	}

	// Train the model
	trainModel(g, prob, labelTensor, w, b)

	// Print the vocabulary
	fmt.Println("Vocabulary:", vocab)

	// Example prediction (for demonstration purposes)
	testText := "the cat is on the mat"
	testInput, _ := preprocess([]string{testText})

	// Run the model to get predictions
	gorgonia.Read(testInput, prob)

	// Print the predicted class
	fmt.Println("Predicted class for test text:", testText)
}
