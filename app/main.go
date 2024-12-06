package main

import (
	"fmt"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"
	// _ "modernc.org/sqlite"
)

const (
	Good bayesian.Class = "good"
	Bad  bayesian.Class = "bad"
)

// label + array of strings
type Data struct {
	Label string   `json:"label"`
	Words []string `json:"text"`
}

// func getTrainData() ([]Data, error) {
// 	return []Data{}, nil
// }

func main() {
	// var err error

	fmt.Println(processText("я тут сижу один в комнате"))
	classifier := bayesian.NewClassifier(Good, Bad)
	classifier.Learn([]string{"я", "хороший", "человек"}, Good)
	classifier.Learn([]string{"я", "плохой", "человек"}, Bad)
	classifier.ConvertTermsFreqToTfIdf()

	_, likely, _ := classifier.LogScores(
		[]string{"ты", "хорошая", "мама"},
	)

	_, class, _ := classifier.ProbScores(
		[]string{"ты", "хорошая", "мама"},
	)

	fmt.Println(likely, class)

	var database Database = Database{path: "../db/main.db", db: nil, rows: nil}
	var model Model
	// var api API

	database.init()

	model.init(database)
	// _, test := model.learn(database, 0.9)
	// model.learn(database)
	testData := model.learnNew(database, 0.8)

	model.classifier.ConvertTermsFreqToTfIdf()
	// fmt.Print(test)

	// testData, err := database.getTestData()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(testData[0].Words)
	// shuffleSlice(testData)
	// fmt.Println(testData[0].Words)

	labels := model.labels
	for i, label := range labels {
		fmt.Println(i, label)
	}
	fmt.Println("-/-/-/-/-")

	// for _, data := range test {
	// 	fmt.Printf(`%s\t`, data.Label)
	// 	for _, word := range data.Words {
	// 		fmt.Printf(" %s", word)
	// 	}
	// 	fmt.Println()
	// }

	var all int = 0
	var correct int = 0

	for _, data := range testData {
		_, likely, _ := model.classifier.LogScores(
			data.Words,
		)

		_, class, _ := model.classifier.ProbScores(
			data.Words,
		)

		if data.Label == string(labels[likely]) || data.Label == string(labels[class]) {
			correct += 1
		}
		// fmt.Println(data.Label, labels[likely], labels[class])
		all += 1

	}
	fmt.Println("res: ", float64(correct)/float64(all))
	fmt.Println("correct / all: ", correct, all)

	fmt.Println(model.classifier.Learned())

	// for i := 0; i < 5; i++ {

	// 	for j, word := range testData[i].Words {
	// 		fmt.Print("(", j, ") ", word, " ")
	// 	}
	// 	fmt.Println("-->")
	// 	fmt.Println(testData[i].Label, "-->")

	// 	scores, likely, _ := model.classifier.LogScores(
	// 		testData[i].Words,
	// 	)

	// 	fmt.Println(scores)
	// 	fmt.Println(likely, labels[likely])

	// 	res, class, _ := model.classifier.ProbScores(
	// 		testData[i].Words,
	// 	)

	// 	fmt.Println(res)
	// 	fmt.Println(class, labels[class])

	// 	fmt.Println("-/-/-/-/-")
	// }

	// probs, likely, _ := model.classifier.ProbScores(
	// 	[]string{"tall", "girl"},
	// )

	// fmt.Println(probs, likely)

	// var input string
	// for {
	// 	fmt.Scan(input)
	// 	model.classifier.Observe(input)
	// }
	// api.init()
	// api.add()
	// api.start()
}
