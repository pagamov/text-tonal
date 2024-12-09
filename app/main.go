package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"

	// _ "modernc.org/sqlite"

	_ "github.com/jmoiron/sqlx"
	"github.com/joho/godotenv" // Load .env file
	_ "github.com/lib/pq"
)

var (
	LogTableSQL = `CREATE TABLE IF NOT EXISTS "log_table" (
        "id" INTEGER PRIMARY KEY,
        "date" TEXT,
        "text" TEXT,
        "label" TEXT,
        "info" TEXT
    );`

	SampleTableSQL = `CREATE TABLE IF NOT EXISTS "sample_table" (
        "id" INTEGER PRIMARY KEY,
        "text_en" TEXT,
        "text_ru" TEXT DEFAULT NULL,
        "label" TEXT,
        "processed" INTEGER DEFAULT 0
    );`

	UsageTableSQL = `CREATE TABLE IF NOT EXISTS "usage_table" (
        "id" INTEGER PRIMARY KEY,
        "word" TEXT NOT NULL,
        "language" TEXT NOT NULL,
        "label" TEXT NOT NULL,
        "usage" INTEGER NOT NULL DEFAULT 0
    );`
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DBNAME"),
		"disable")

	pgDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgDB.Close()

	err = pgDB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to the database!")

	_, err = pgDB.Exec(`	DO $$ 
DECLARE 
    r RECORD; 
BEGIN 
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP 
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; 
    END LOOP; 
END $$;`)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("old data deleted")

	// Create tables in PostgreSQL
	_, err = pgDB.Exec(LogTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Log done")
	_, err = pgDB.Exec(SampleTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Sample done")
	_, err = pgDB.Exec(UsageTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Usage done")

	// Connect to SQLite
	sqliteDB, err := sql.Open("sqlite3", "../db/main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteDB.Close()

	// Transfer data from Log_table
	transferLogData(sqliteDB, pgDB)

	log.Print("Log done")

	// Transfer data from Sample_table
	transferSampleData(sqliteDB, pgDB)

	log.Print("Sample done")

	// Transfer data from Usage_table
	transferUsageData(sqliteDB, pgDB)

	log.Print("Usage done")
	fmt.Println("Data transfer completed successfully!")

	log.Fatal()

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
	database.replaceLabels()

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
	// api.addMethod()
	// api.start()
}
