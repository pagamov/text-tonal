package main

import (
	"fmt"
	"log"
	"strings"

	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func createBagOfWords(text string) map[string]int {
	words := strings.Fields(text)
	bow := make(map[string]int)

	for _, word := range words {
		bow[word]++
	}

	return bow
}

// Define the structure of the neural network
func createModel(g *gorgonia.ExprGraph, inputSize, numClasses int) (*gorgonia.Node, *gorgonia.Node) {
	// Input layer
	x := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(-1, inputSize), gorgonia.WithName("x"))

	// Hidden layer
	w0 := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(inputSize, 10), gorgonia.WithName("w0"), gorgonia.WithValue(tensor.New(tensor.WithShape(inputSize, 10), tensor.WithBacking(make([]float32, inputSize*10)))))
	b0 := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(1, 10), gorgonia.WithName("b0"), gorgonia.WithValue(tensor.New(tensor.WithShape(1, 10), tensor.WithBacking(make([]float32, 10)))))
	h0 := gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(x, w0)), b0))
	h0Act := gorgonia.Must(gorgonia.Rectify(h0)) // Activation function

	// Output layer
	w1 := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(10, numClasses), gorgonia.WithName("w1"), gorgonia.WithValue(tensor.New(tensor.WithShape(10, numClasses), tensor.WithBacking(make([]float32, 10*numClasses)))))
	b1 := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(1, numClasses), gorgonia.WithName("b1"), gorgonia.WithValue(tensor.New(tensor.WithShape(1, numClasses), tensor.WithBacking(make([]float32, numClasses)))))
	logits := gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(h0Act, w1)), b1))
	prob := gorgonia.Must(gorgonia.SoftMax(logits))

	return x, prob
}

func main() {
	g := gorgonia.NewGraph()

	// Define input size and number of classes
	inputSize := 100 // Example input size (e.g., word embeddings)
	numClasses := 5  // Example number of emotion classes

	x, prob := createModel(g, inputSize, numClasses)

	// Example input data (replace with actual data)
	inputData := tensor.New(tensor.WithShape(1, inputSize), tensor.WithBacking(make([]float32, inputSize)))
	gorgonia.WithValue(inputData)(x)

	// Create a VM to run the graph
	vm := gorgonia.NewTapeMachine(g)

	// Run the model
	if err := vm.RunAll(); err != nil {
		log.Fatal(err)
	}

	// Output the probabilities
	fmt.Println("Predicted probabilities:", prob.Value())
}
