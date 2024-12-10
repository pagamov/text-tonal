package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
)

func main() {
	r := gin.Default()

	r.POST("/predict", func(c *gin.Context) {
		var json struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&json ); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Here you would preprocess the text and run it through the model
		// For demonstration, we'll return a dummy response
		c.JSON(http.StatusOK, gin.H{"predicted_class": "happy"})
	})

	r.Run(":8080") // Start the server on port 8080
}


func import_from_old() {
// Connect to the main database
db, err := sql.Open("sqlite3", "db/main.db")
if err != nil {
	log.Fatal(err)
}
defer db.Close()

// Create Log_table
_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "Log_table" (
	"id" INTEGER PRIMARY KEY,
	"date" TEXT,
	"text" TEXT,
	"label" TEXT,
	"info" TEXT
);`)
if err != nil {
	log.Fatal(err)
}

// Create Sample_table
_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "Sample_table" (
	"id" INTEGER PRIMARY KEY,
	"text_en" TEXT,
	"text_ru TEXT DEFAULT NULL,
	"label" TEXT,
	"processed" INTEGER DEFAULT 0
);`)
if err != nil {
	log.Fatal(err)
}

// Clear Sample_table
_, err = db.Exec(`DELETE FROM "Sample_table";`)
if err != nil {
	log.Fatal(err)
}

// Insert data from old databases
var g errgroup.Group
for i := 0; i < 10; i++ {
	i := i // capture range variable
	g.Go(func() error {
		conOld, err := sql.Open("sqlite3", filepath.Join("homework/data/db", fmt.Sprintf("mydatabase_%d.db", i)))
		if err != nil {
			return err
		}
		defer conOld.Close()

		rows, err := conOld.Query("SELECT * FROM emotions")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var textEn, textRu, label string
			var processed int
			if err := rows.Scan(&id, &textEn, &textRu, &label, &processed); err != nil {
				return err
			}
			_, err = db.Exec("INSERT INTO \"Sample_table\" (text_en, text_ru, label, processed) VALUES (?, ?, ?, ?)", textEn, textRu, label, 0)
			if err != nil {
				return err
			}
		}
		return rows.Err()
	})
}

if err := g.Wait(); err != nil {
	log.Fatal(err)
}

// Create Usage_table
_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "Usage_table" (
	"id" INTEGER PRIMARY KEY,
	"word" TEXT NOT NULL,
	"language" TEXT NOT NULL,
	"label" TEXT NOT NULL,
	"usage" INTEGER NOT NULL DEFAULT 0
);`)
if err != nil {
	log.Fatal(err)
}

// Clear Usage_table
_, err = db.Exec(`DELETE FROM "Usage_table";`)
if err != nil {
	log.Fatal(err)
}

// Select rows from Sample_table
rows, err := db.Query(`SELECT * FROM "Sample_table" WHERE "text_ru" IS NOT NULL;`)
if err != nil {
	log.Fatal(err)
}
defer rows.Close()

allUsage := make(map[string]map[string]int)

for rows.Next() {
	var id int
	var textEn, textRu, label string
	var processed int
	if err := rows.Scan(&id, &textEn, &textRu, &label, &processed); err != nil {
		log.Fatal(err)
	}

	if _, exists := allUsage[label]; !exists {
		allUsage[label] = make(map[string]int)
	}

	// Process words
	words := regexp.MustCompile(`\s+`).Split(textEn, -1)
	for _, word := range words {
		newWord := regexp.MustCompile(`[^a-zA-Z]`).ReplaceAllString(word, "")
		if newWord != "" {
			allUsage[label][newWord]++
		}
	}
}

// Insert usage data into Usage_table
for label, words := range allUsage {
	for word, count := range words {
		_, err = db.Exec("INSERT INTO \"Usage_table\" (word, language, label, usage) VALUES (?, ?, ?, ?)", word, "ru", label, count)
		if err != nil {
			log.Fatal(err)
		}
	}
}

fmt.Println("Data processing complete.")
}


type Database struct {
	path string
	db   *sql.DB
	rows *sql.Rows
}

func transferLogData(sqliteDB, pgDB *sql.DB) {
	rows, err := sqliteDB.Query("SELECT id, date, text, label, info FROM Log_table")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var date, text, label, info string
		if err := rows.Scan(&id, &date, &text, &label, &info); err != nil {
			log.Fatal(err)
		}

		_, err = pgDB.Exec("INSERT INTO log_table (id, date, text, label, info) VALUES ($1, $2, $3, $4, $5)",
			id, date, text, label, info)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func transferSampleData(sqliteDB, pgDB *sql.DB) {
	rows, err := sqliteDB.Query("SELECT id, text_en, text_ru, label, processed FROM Sample_table")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var textEn, label string
		var textRu sql.NullString
		var processed int

		if err := rows.Scan(&id, &textEn, &textRu, &label, &processed); err != nil {
			log.Fatal(err)
		}

		_, err = pgDB.Exec("INSERT INTO sample_table (id, text_en, text_ru, label, processed) VALUES ($1, $2, $3, $4, $5)",
			id, textEn, textRu, label, processed)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func transferUsageData(sqliteDB, pgDB *sql.DB) {
	rows, err := sqliteDB.Query("SELECT id, word, language, label, usage FROM Usage_table")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var word, language, label string
		var usage int
		if err := rows.Scan(&id, &word, &language, &label, &usage); err != nil {
			log.Fatal(err)
		}

		_, err = pgDB.Exec("INSERT INTO usage_table (id, word, language, label, usage) VALUES ($1, $2, $3, $4, $5)",
			id, word, language, label, usage)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// setup all db tables
func (database *Database) init() {
	var db *sql.DB
	var err error
	var path string = database.path

	var Log_table string = `CREATE TABLE IF NOT EXISTS "Log_table" (
		"id"	INTEGER,
		"date"	TEXT,
		"text"	TEXT,
		"label"	TEXT,
		"info"	TEXT,
		PRIMARY KEY("id")
	);`
	var Sample_table string = `
	CREATE TABLE IF NOT EXISTS "Sample_table" (
		"id"	INTEGER,
		"text_en"	TEXT,
		"text_ru"	TEXT DEFAULT NULL,
		"label"	TEXT,
		"processed"	INTEGER DEFAULT 0,
		PRIMARY KEY("id")
	);`
	var Usage_table string = `
	CREATE TABLE IF NOT EXISTS "Usage_table" (
		"id"	INTEGER,
		"word"	TEXT NOT NULL,
		"language"	TEXT NOT NULL,
		"label"	TEXT NOT NULL,
		"usage"	INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY("id")
	);`

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(Log_table); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(Sample_table); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(Usage_table); err != nil {
		log.Fatal(err)
	}
	log.Print("db inited")
}

func (database *Database) getTestData() ([]Data, error) {
	var err error
	var res []Data = []Data{}
	var db *sql.DB
	var rows *sql.Rows
	var querry string = `
		SELECT label, text_en from Sample_table where text_en is not NULL;
	`
	var path string = database.path
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

// we get label and return list of all words with this label
func (database *Database) getWordsByLabel(label string) ([]string, error) {
	var res []string
	var db *sql.DB
	var querry string
	var rows *sql.Rows
	var err error
	var word string

	db, err = sql.Open("sqlite3", database.path)
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
func (database *Database) getUsage(labels []bayesian.Class) ([]Data, error) {
	var label bayesian.Class
	var res []Data = []Data{}
	var err error
	var words []string

	for _, label = range labels {
		words, err = database.getWordsByLabel(string(label))
		if err != nil {
			return []Data{}, err
		}
		res = append(res, Data{Label: string(label), Words: words})
	}
	return res, nil
}

// get all labels of text from db
func (database *Database) getLabels() ([]bayesian.Class, error) {
	var res []bayesian.Class
	var db *sql.DB
	var rows *sql.Rows
	var err error
	var label string
	var path string = database.path
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

func (database *Database) replaceLabels() {
	var db *sql.DB
	var err error
	var path string = database.path

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	q1 := `UPDATE Sample_table SET label = 'neutral' WHERE label = 'empty' or 'relief' or 'surprise' or 'worry' or 'boredom';`
	q2 := `UPDATE Sample_table SET label = 'good' WHERE label = 'enthusiasm' or 'fun' or 'happiness' or 'joy' or 'love';`
	q3 := `UPDATE Sample_table SET label = 'bad' WHERE label = 'anger' or 'fear' or 'hate' or 'sadness';`

	if _, err := db.Exec(q1); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q2); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q3); err != nil {
		log.Fatal(err)
	}

	q4 := `UPDATE Usage_table SET label = 'neutral' WHERE label = 'empty' or 'relief' or 'surprise' or 'worry' or 'boredom';`
	q5 := `UPDATE Usage_table SET label = 'good' WHERE label = 'enthusiasm' or 'fun' or 'happiness' or 'joy' or 'love';`
	q6 := `UPDATE Usage_table SET label = 'bad' WHERE label = 'anger' or 'fear' or 'hate' or 'sadness';`

	if _, err := db.Exec(q4); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q5); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q6); err != nil {
		log.Fatal(err)
	}

}

// for line of text return splitted []string
// of rus words without trash
func processText(text string) []string {
	var words []string
	// lower text
	var lower string = strings.ToLower(text)
	// create re regex filter
	re := regexp.MustCompile(`[^a-z\s]+`)
	// filter words
	cleaned := re.ReplaceAllString(lower, "")
	// split string by " "
	words = strings.Split(cleaned, " ")
	return words

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

// shuffleSlice shuffles the elements of a slice.
func shuffleSlice(slice []Data) {
	rand.Seed(uint64(time.Now().UnixNano())) // Seed the random number generator
	for i := range slice {
		j := rand.Intn(i + 1)                   // Generate a random index
		slice[i], slice[j] = slice[j], slice[i] // Swap elements
	}
}

func (model *Model) learn(database Database) {
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

func (model *Model) learnNew(database Database, ratio float64) []Data {
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
	return test
}


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

type API struct {
	router *gin.Engine
}

func (api *API) init() {
	api.router = gin.Default()
}

func (api *API) addMethod() {
	api.router.POST("/analyze", analyze)
	api.router.GET("/statistics", statistics)
}

func (api *API) start(port int) {
	api.router.Run(fmt.Sprintf(":%s", string(port)))
}

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
