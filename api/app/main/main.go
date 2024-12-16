package main

import (
	"api/db"
	"api/model"
	"api/router"
	"fmt"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"

	// _ "modernc.org/sqlite"
	_ "github.com/jmoiron/sqlx" // Load .env file
	_ "github.com/lib/pq"
)

var (
	TextModel      model.Model
	api            router.Router
	databaseSqlite db.DatabaseSQLite
)

func main() {

	databaseSqlite = *db.CreateDatabaseSQLite("../db/main.db")
	databaseSqlite.Init()
	databaseSqlite.ReplaceLabels()
	databaseSqlite.PrintLabels()

	TextModel.Init(databaseSqlite)

	testData := TextModel.LearnWithBag(databaseSqlite, 0.8, true)

	TextModel.Test(testData)

	api.Init()
	api.AddMethod()
	api.Start("8080")
}

func testSimpleModel() {

	var Good bayesian.Class = "good"
	var Bad bayesian.Class = "bad"

	fmt.Println(db.ProcessText("я тут сижу один в комнате"))
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
