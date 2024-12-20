package model

import (
	"fmt"
	"log"
	"time"

	// "github.com/fogfish/word2vec"
	"github.com/jbrukh/bayesian"
	"golang.org/x/exp/rand"

	"api/data"
	"api/db"
)

type Model struct {
	classifier *bayesian.Classifier
	labels     []bayesian.Class

	// llm      *openai.LLM
	// embedder *embeddings.EmbedderImpl

	// w2v word2vec.Model
}

func (model *Model) Test(testData []data.Data) {

	var all int = 0
	var correct int = 0

	for _, data := range testData {
		stringSlice1 := make([]string, len(data.Vec))
		for i, v := range data.Vec {
			stringSlice1[i] = fmt.Sprintf("%f", v) // You can format as needed
		}

		_, likely, _ := model.classifier.LogScores(
			data.Words,
		)

		_, class, _ := model.classifier.ProbScores(
			data.Words,
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

// type DatabaseSQLite struct {
// 	path string
// 	db   *sql.DB
// 	rows *sql.Rows
// }

// get all labels from DB

func (model *Model) Init(database db.DatabaseSQLite) {
	var labels []bayesian.Class
	var err error

	labels, err = database.GetLabels()
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

	// model.w2v, err = word2vec.Load("wap-v300w5e5s1h005-en.bin", 300)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _ = embs
	// for i, l := range labels {
	// 	fmt.Println(i, l)
	// }
	model.classifier = bayesian.NewClassifier(labels...)
	model.labels = labels
}

// shuffleSlice shuffles the elements of a slice.
func shuffleSlice(slice []data.Data) {
	rand.Seed(uint64(time.Now().UnixNano())) // Seed the random number generator
	for i := range slice {
		j := rand.Intn(i + 1)                   // Generate a random index
		slice[i], slice[j] = slice[j], slice[i] // Swap elements
	}
}

// removeFrequentWords принимает массив массивов слов и удаляет самые частые слова
func removeFrequentWords(data []data.Data, threshold int) []data.Data {
	wordCount := make(map[string]int)

	// Подсчет частоты слов
	for _, words := range data {
		for _, word := range words.Words {
			wordCount[word]++
		}
	}

	// Создание множества самых частых слов
	frequentWords := make(map[string]struct{})
	for word, count := range wordCount {
		if count >= threshold {
			frequentWords[word] = struct{}{}
		}
	}

	// Создание нового массива без самых частых слов
	for i, words := range data {
		var filteredWords []string
		for _, word := range words.Words {
			if _, exists := frequentWords[word]; !exists {
				filteredWords = append(filteredWords, word)
			}
		}
		data[i].Words = filteredWords
		// result = append(result, filteredWords)
	}

	log.Println("removeFrequentWords", frequentWords)

	return data
}

func (model *Model) learn(database db.DatabaseSQLite) {
	var class data.Data
	data, err := database.GetUsage(model.labels)
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

func (model *Model) LearnNew(database db.DatabaseSQLite, ratio float64, convert bool) []data.Data {
	data, err := database.GetTestData()
	if err != nil {
		log.Fatal(err)
	}

	shuffleSlice(data)

	data = removeFrequentWords(data, 100)

	// Calculate the number of elements to take (ratio%)
	ninetyPercentCount := int(ratio * float64(len(data)))

	// Take the first 90% of the shuffled array
	train := data[:ninetyPercentCount]
	test := data[ninetyPercentCount:]

	for _, t := range train {
		model.classifier.Learn(t.Words, bayesian.Class(t.Label))
		// log.Println(t.Words, bayesian.Class(t.Label))
	}
	log.Println("model learned new")

	if convert {
		model.classifier.ConvertTermsFreqToTfIdf()
		log.Println("ConvertTermsFreqToTfIdf completed")
	}

	return test
}

func (model *Model) embed(data []data.Data) []data.Data {
	for i := range data {
		data[i].Vec = make([]float32, 300)
		// model.w2v.Embedding(strings.Join(data[i].Words, " "), data[i].Vec)
	}
	return data
}

func (model *Model) LearnWithBag(database db.DatabaseSQLite, ratio float64, convert bool) []data.Data {
	data, err := database.GetTestData()
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
