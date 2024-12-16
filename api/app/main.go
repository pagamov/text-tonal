package main

import (
	"fmt"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"

	// _ "modernc.org/sqlite"
	_ "github.com/jmoiron/sqlx" // Load .env file
	_ "github.com/lib/pq"
)

const ()

// label + array of strings
type Data struct {
	Label string    `json:"label"`
	Words []string  `json:"text"`
	Vec   []float32 `json:"vec"`
}

func main() {

	var databaseSqlite DatabaseSQLite = DatabaseSQLite{path: "../db/main.db", db: nil, rows: nil}

	var model Model

	var api API

	databaseSqlite.init()
	databaseSqlite.replaceLabels()
	databaseSqlite.printLabels()

	model.init(databaseSqlite)

	testData := model.learnWithBag(databaseSqlite, 0.8, true)

	model.test(testData)

	api.init()
	api.addMethod()
	api.start("8080")
}

func testSimpleModel() {

	var Good bayesian.Class = "good"
	var Bad bayesian.Class = "bad"

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
}
