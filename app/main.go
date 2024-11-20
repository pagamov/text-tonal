package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
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

// can be multiple labels for one word
type Info struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// for word we got N info marks for each label
type Word struct {
	Word string `json:"word"`
	Info []Info `json:"info"`
}

type Analyz struct {
	Count int64  `json:"count"`
	Label string `json:"label"`
	Words []Word `json:"words"`
}

type Statistics struct {
	Date  string `json:"date"`
	Text  string `json:"text"`
	Count int64  `json:"count"`
	Label string `json:"label"`
	Words []Word `json:"words"`
}

// label + array of strings
type Data struct {
	Label string   `json:"label"`
	Words []string `json:"text"`
}

// const (
// 	Good bayesian.Class = "Good"
// 	Bad  bayesian.Class = "Bad"
// )

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

func analyze(c *gin.Context) {
	// 	POST API/analyze?text=some text to parse

	// 	RES =  {
	//         "count" : "Number of words : Int64",
	//         "label" : "soft max label of text : String",
	//         "words" : [
	//             {
	//                 "word" : "word itself : String",
	//                 "info" : [
	//                     {
	//                         "label" : "some label from learning labels : String",
	//                         "value" : "percentage : Int8"
	//                     }
	//                 ]
	//             }
	//         ]
	// }

	res := Analyz{
		Count: 10,
		Label: "label",
		Words: []Word{
			{
				Word: "word",
				Info: []Info{
					{Label: "label", Value: 10},
				},
			},
		},
	}

	c.IndentedJSON(http.StatusOK, res)
}

func statistics(c *gin.Context) {
	// GET API/statistics?date_begin=“dd.mm.yyyy”&date_end==“dd.mm.yyyy”
	// RES =  [{
	// 	"date" : "date of request : Date",
	// 	"text" : "text : String",
	// 	"count" : "Number of words : Int64",
	// 			"label" : "soft max label of text : String",
	// 			"words" : [
	// 				{  "word" : "word itself : String",
	// 					"info" : [{
	// 							"label" : "some label from learning labels : String",
	// 							"value" : "percentage : Int8"
	// 						}]
	// 				}
	// 			]
	// 	}]

	var res []Statistics = []Statistics{
		{
			Date:  "01/01/1977 14:20:00",
			Text:  "Some text",
			Count: 10,
			Label: "label",
			Words: []Word{
				{
					Word: "word",
					Info: []Info{
						{
							Label: "label",
							Value: 0,
						},
					},
				},
			},
		},
	}

	c.IndentedJSON(http.StatusOK, res)
}

// setup all db tables
func setupDB() error {
	var db *sql.DB
	var err error
	var path string = path_db

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(Log_table); err != nil {
		return err
	}
	if _, err := db.Exec(Sample_table); err != nil {
		return err
	}
	if _, err := db.Exec(Usage_table); err != nil {
		return err
	}
	return nil
}

// we get label and return list of all words with this label
func getWordsByLabel(label string) ([]string, error) {
	var res []string
	var db *sql.DB
	var querry string
	var rows *sql.Rows
	var err error
	var word string

	db, err = sql.Open("sqlite3", path_db)
	if err != nil {
		return []string{}, err
	}
	defer db.Close()
	querry = fmt.Sprintf(`SELECT word from Usage_table where label is "%s";`, label)
	rows, err = db.Query(querry)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&word)
		if err != nil {
			return []string{}, err
		}
		res = append(res, word)
	}
	return res, nil
}

// get Data struct with map of labels and corresponding list of words
func getUsage(labels []bayesian.Class) ([]Data, error) {
	var label bayesian.Class
	var res []Data = []Data{}
	var err error
	var words []string

	for _, label = range labels {
		words, err = getWordsByLabel(string(label))
		if err != nil {
			return []Data{}, err
		}
		res = append(res, Data{Label: string(label), Words: words})
	}
	return []Data{}, nil
}

// get all labels of text from db
func getLabels() ([]bayesian.Class, error) {
	var res []bayesian.Class
	var db *sql.DB
	var rows *sql.Rows
	var err error
	var label string
	var path string = path_db
	var getLabels string = `
		SELECT label FROM Usage_table GROUP by label ORDER BY label ASC;
	`

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return []bayesian.Class{}, err
	}
	defer db.Close()

	rows, err = db.Query(getLabels)
	if err != nil {
		return []bayesian.Class{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&label)
		if err != nil {
			return []bayesian.Class{}, err
		}
		res = append(res, bayesian.Class(label))
	}
	return res, nil
}

// for line of text return splitted []string
// of rus words without trash
func processText(text string) []string {
	var words []string

	// lower text
	var lower string = strings.ToLower(text)
	// create re regex filter
	re := regexp.MustCompile(`[^а-яё]`)
	// filter words
	cleaned := re.ReplaceAllString(lower, "")
	// split string by " "
	words = strings.Split(cleaned, " ")

	return words
}

// func getTrainData() ([]Data, error) {
// 	return []Data{}, nil
// }

func getTestData() ([]Data, error) {
	var err error
	var res []Data = []Data{}
	var db *sql.DB
	var rows *sql.Rows
	var querry string = `
		SELECT label, text_ru from Sample_table where text_ru is not NULL;
	`
	var path string = path_db
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return []Data{}, err
	}
	defer db.Close()

	rows, err = db.Query(querry)
	if err != nil {
		return []Data{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		var text string
		var processedText []string
		err = rows.Scan(&label, &text)
		if err != nil {
			return []Data{}, err
		}
		processedText = processText(text)
		res = append(res, Data{Label: label, Words: processedText})
	}
	return res, nil
}

func main() {
	var err error
	var labels []bayesian.Class
	var index int
	var label bayesian.Class
	var router *gin.Engine

	err = setupDB()
	if err != nil {
		log.Fatal(err)
	}

	labels, err = getLabels()
	if err != nil {
		log.Fatal(err)
	}

	for index, label = range labels {
		fmt.Printf("Label %d: %s\n", index, label)
	}

	data, err := getUsage(labels)
	if err != nil {
		log.Fatal(err)
	}

	classifier := bayesian.NewClassifier(labels...)

	var class Data
	for _, class = range data {
		classifier.Learn(class.Words, bayesian.Class(class.Label))
	}

	testData, err := getTestData()
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

	scores, likely, _ := classifier.LogScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(scores, likely)

	probs, likely, _ := classifier.ProbScores(
		[]string{"tall", "girl"},
	)

	fmt.Println(probs, likely)

	router = gin.Default()
	router.POST("/analyze", analyze)
	router.GET("/statistics", statistics)
	router.Run(":8080")
}
