package main

import (
	"database/sql"
	"fmt"
	"log"

	// "net/http"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Log_table string = `CREATE TABLE IF NOT EXISTS "Log_table" (
		"id"	INTEGER,
		"date"	TEXT,
		"text"	TEXT,
		"label"	TEXT,
		"info"	TEXT,
		PRIMARY KEY("id")
	);`
	Sample_table string = `
	CREATE TABLE IF NOT EXISTS "Sample_table" (
		"id"	INTEGER,
		"text_en"	TEXT,
		"text_ru"	TEXT DEFAULT NULL,
		"label"	TEXT,
		"processed"	INTEGER DEFAULT 0,
		PRIMARY KEY("id")
	);`
	Usage_table string = `
	CREATE TABLE IF NOT EXISTS "Usage_table" (
		"id"	INTEGER,
		"word"	TEXT NOT NULL,
		"language"	TEXT NOT NULL,
		"label"	TEXT NOT NULL,
		"usage"	INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY("id")
	);`
	path_db string = "../db/main.db"
)

const (
	Good bayesian.Class = "Good"
	Bad  bayesian.Class = "Bad"
)

// album represents data about a record album.
// type album struct {
// 	ID     string  `json:"id"`
// 	Title  string  `json:"title"`
// 	Artist string  `json:"artist"`
// 	Price  float64 `json:"price"`
// }

// albums slice to seed record album data.
// var albums = []album{
// 	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
// 	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
// 	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
// }

// func getAlbums(c *gin.Context) {
// 	c.IndentedJSON(http.StatusOK, albums)
// }

func setupDB() {
	db, err := sql.Open("sqlite3", path_db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(Log_table); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	if _, err := db.Exec(Sample_table); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	if _, err := db.Exec(Usage_table); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func getLabels() []bayesian.Class {

	var res []bayesian.Class
	db, err := sql.Open("sqlite3", path_db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	getLabels := `
	SELECT label FROM Usage_table GROUP by label;
	`
	rows, err := db.Query(getLabels)
	if err != nil {
		log.Fatalf("Error getting labels: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		err = rows.Scan(&label)
		if err != nil {
			log.Fatalf("Error scanning label: %v", err)
		}
		res = append(res, bayesian.Class(label))
	}
	return res
}

func main() {
	setupDB()

	var labels []bayesian.Class = getLabels()

	for index, label := range labels {
		fmt.Printf("Label %d: %s\n", index, label)
	}

	classifier := bayesian.NewClassifier(Good, Bad)
	goodStuff := []string{"tall", "rich", "handsome"}
	badStuff := []string{"tall", "poor", "smelly", "ugly"}
	classifier.Learn(goodStuff, Good)
	classifier.Learn(badStuff, Bad)

	scores, likely, _ := classifier.LogScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(scores, likely)

	probs, likely, _ := classifier.ProbScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(probs, likely)

	// router := gin.Default()
	// router.GET("/albums", getAlbums)
	// router.Run(":8080")

}
