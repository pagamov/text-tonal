package main

import (
	"log"
	"time"

	"github.com/jbrukh/bayesian"
	"golang.org/x/exp/rand"
)

type Model struct {
	classifier *bayesian.Classifier
	labels     []bayesian.Class
}

// get all labels from DB

func (model *Model) init(database DatabaseSQLite) {
	var labels []bayesian.Class
	var err error

	labels, err = database.getLabels()
	if err != nil {
		log.Fatal(err)
	}
	// for i, l := range labels {
	// 	fmt.Println(i, l)
	// }
	model.classifier = bayesian.NewClassifier(labels...)
	model.labels = labels
}

// shuffleSlice shuffles the elements of a slice.
func shuffleSlice(slice []Data) {
	rand.Seed(uint64(time.Now().UnixNano())) // Seed the random number generator
	for i := range slice {
		j := rand.Intn(i + 1)                   // Generate a random index
		slice[i], slice[j] = slice[j], slice[i] // Swap elements
	}
}

func (model *Model) learn(database DatabaseSQLite) {
	var class Data
	data, err := database.getUsage(model.labels)
	if err != nil {
		log.Fatal(err)
	}

	shuffleSlice(data)

	// Calculate the number of elements to take (90%)
	// ninetyPercentCount := int(1 * float64(len(data)))

	// Take the first 90% of the shuffled array
	// train := data[:ninetyPercentCount]
	// test := data[ninetyPercentCount:]

	for _, class = range data {
		model.classifier.Learn(class.Words, bayesian.Class(class.Label))
	}
	log.Println("model learned")
}

func (model *Model) learnNew(database DatabaseSQLite, ratio float64) []Data {
	data, err := database.getTestData()
	if err != nil {
		log.Fatal(err)
	}

	shuffleSlice(data)

	// Calculate the number of elements to take (90%)
	ninetyPercentCount := int(ratio * float64(len(data)))

	// Take the first 90% of the shuffled array
	train := data[:ninetyPercentCount]
	test := data[ninetyPercentCount:]

	for _, t := range train {
		model.classifier.Learn(t.Words, bayesian.Class(t.Label))
	}
	log.Println("model learned new")
	return test
}
