package main

import (
	"fmt"
	"log"

	// _ "github.com/mattn/go-sqlite3"

	_ "modernc.org/sqlite"
)

const ()

// label + array of strings
type Data struct {
	Label string   `json:"label"`
	Words []string `json:"text"`
}

// func getTrainData() ([]Data, error) {
// 	return []Data{}, nil
// }

func main() {
	var err error

	var database Database = Database{path: "../db/main.db", db: nil, rows: nil}
	var model Model
	var api API

	database.init()

	model.init(database)
	model.learn(database)

	testData, err := database.getTestData()
	if err != nil {
		log.Fatal(err)
	}

	for _, data := range testData {
		fmt.Printf(`%s\t`, data.Label)
		for _, word := range data.Words {
			fmt.Printf(" %s", word)
		}
		fmt.Println()
	}

	scores, likely, _ := model.classifier.LogScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(scores, likely)

	probs, likely, _ := model.classifier.ProbScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(probs, likely)

	api.init()
	api.add()
	api.start()
}
