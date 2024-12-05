package main

import (
	"log"

	"github.com/jbrukh/bayesian"
)

type Model struct {
	classifier *bayesian.Classifier
	labels     []bayesian.Class
}

// get all labels from DB

func (model *Model) init(database Database) {
	var labels []bayesian.Class
	var err error

	labels, err = database.getLabels()
	if err != nil {
		log.Fatal(err)
	}
	model.classifier = bayesian.NewClassifier(labels...)
	model.labels = labels
}

func (model *Model) learn(database Database) {
	var class Data
	data, err := database.getUsage(model.labels)
	if err != nil {
		log.Fatal(err)
	}

	for _, class = range data {
		model.classifier.Learn(class.Words, bayesian.Class(class.Label))
	}
}
