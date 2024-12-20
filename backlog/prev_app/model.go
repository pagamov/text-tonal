package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fogfish/word2vec"
	"github.com/jbrukh/bayesian"
	"golang.org/x/exp/rand"
)

type Model struct {
	classifier *bayesian.Classifier
	labels     []bayesian.Class

	// llm      *openai.LLM
	// embedder *embeddings.EmbedderImpl

	w2v word2vec.Model
}

func (model *Model) test(testData []Data) {

	var all int = 0
	var correct int = 0

	for _, data := range testData {
		stringSlice1 := make([]string, len(data.Vec))
		for i, v := range data.Vec {
			stringSlice1[i] = fmt.Sprintf("%f", v) // You can format as needed
		}

		_, likely, _ := model.classifier.LogScores(
			stringSlice1,
		)

		_, class, _ := model.classifier.ProbScores(
			stringSlice1,
		)

		if data.Label == string(model.labels[likely]) || data.Label == string(model.labels[class]) {
			correct += 1
		}
		all += 1
	}
	log.Println("res: ", float64(correct)/float64(all))
	log.Println("correct / all: ", correct, all)
	log.Println("model.Learned: ", model.classifier.Learned())
}

// get all labels from DB

func (model *Model) init(database DatabaseSQLite) {
	var labels []bayesian.Class
	var err error

	labels, err = database.getLabels()
	if err != nil {
		log.Fatal(err)
	}

	// model.llm, err = openai.New()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// model.embedder, err = embeddings.NewEmbedder(model.llm)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// docs := []string{"doc 1", "another doc"}
	// embs, err := model.embedder.EmbedDocuments(context.Background(), docs)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	model.w2v, err = word2vec.Load("wap-v300w5e5s1h005-en.bin", 300)
	if err != nil {
		log.Fatal(err)
	}

	// _ = embs
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

func (model *Model) learnNew(database DatabaseSQLite, ratio float64, convert bool) []Data {
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

	if convert {
		model.classifier.ConvertTermsFreqToTfIdf()
		log.Println("ConvertTermsFreqToTfIdf completed")
	}

	return test
}

func (model *Model) embed(data []Data) []Data {
	for i := range data {
		data[i].Vec = make([]float32, 300)
		model.w2v.Embedding(strings.Join(data[i].Words, " "), data[i].Vec)
	}
	return data
}

func (model *Model) learnWithBag(database DatabaseSQLite, ratio float64, convert bool) []Data {
	data, err := database.getTestData()
	if err != nil {
		log.Fatal(err)
	}

	shuffleSlice(data)

	data = model.embed(data)

	// Calculate the number of elements to take (90%)
	ninetyPercentCount := int(ratio * float64(len(data)))

	// Take the first 90% of the shuffled array
	train := data[:ninetyPercentCount]
	test := data[ninetyPercentCount:]

	for _, t := range train {
		stringSlice1 := make([]string, len(t.Vec))
		for i, v := range t.Vec {
			stringSlice1[i] = fmt.Sprintf("%f", v) // You can format as needed
		}
		model.classifier.Learn(stringSlice1, bayesian.Class(t.Label))
	}
	log.Println("model learned new")

	if convert {
		model.classifier.ConvertTermsFreqToTfIdf()
		log.Println("ConvertTermsFreqToTfIdf completed")
	}

	return test
}
