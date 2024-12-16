package db

import (
	"api/data"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"
	// _ "modernc.org/sqlite"
)

const (
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

// Database class

type DatabaseSQLite struct {
	path string
	db   *sql.DB
	rows *sql.Rows
}

func CreateDatabaseSQLite(path string) *DatabaseSQLite {
	var res DatabaseSQLite = DatabaseSQLite{path: path, db: nil, rows: nil}
	return &res
}

func (db *DatabaseSQLite) PrintLabels() {
	labels, _ := db.GetLabels()
	for i, label := range labels {
		log.Println(i, label)
	}
}

func TransferLogData(sqliteDB, pgDB *sql.DB) {
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

func TransferSampleData(sqliteDB, pgDB *sql.DB) {
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

func TransferUsageData(sqliteDB, pgDB *sql.DB) {
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
func (database *DatabaseSQLite) Init() {
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

func (database *DatabaseSQLite) GetTestData() ([]data.Data, error) {
	var err error
	var res []data.Data = []data.Data{}
	var db *sql.DB
	var rows *sql.Rows
	var querry string = `
		SELECT label, text_en from Sample_table where text_en is not NULL;
	`
	var path string = database.path
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return []data.Data{}, err
	}
	defer db.Close()

	rows, err = db.Query(querry)
	if err != nil {
		return []data.Data{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		var text string
		var processedText []string
		err = rows.Scan(&label, &text)
		if err != nil {
			return []data.Data{}, err
		}
		processedText = ProcessText(text)
		res = append(res, data.Data{Label: label, Words: processedText})
	}
	return res, nil
}

// we get label and return list of all words with this label
func (database *DatabaseSQLite) getWordsByLabel(label string) ([]string, error) {
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
func (database *DatabaseSQLite) GetUsage(labels []bayesian.Class) ([]data.Data, error) {
	var label bayesian.Class
	var res []data.Data = []data.Data{}
	var err error
	var words []string

	for _, label = range labels {
		words, err = database.getWordsByLabel(string(label))
		if err != nil {
			return []data.Data{}, err
		}
		res = append(res, data.Data{Label: string(label), Words: words})
	}
	return res, nil
}

// get all labels of text from db
func (database *DatabaseSQLite) GetLabels() ([]bayesian.Class, error) {
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

// func CreateCon(path string) (Database, error) {
// 	db, err := sql.Open("sqlite3", path)
// 	if err != nil {
// 		return Database{}, err
// 	}
// 	var database = Database{path: path, db: db, rows: nil}
// 	return database, nil
// }

// func CloseCon(database *Database) error {
// 	var err error
// 	if database.db != nil {
// 		err = database.db.Close()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (database *DatabaseSQLite) ReplaceLabels() {
	var db *sql.DB
	var err error
	var path string = database.path

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	q1 := `UPDATE Sample_table SET label = 'neutral' WHERE label in ('empty' , 'relief' , 'surprise' , 'worry' , 'boredom');`
	q2 := `UPDATE Sample_table SET label = 'good' WHERE label in ('enthusiasm' , 'fun' , 'happiness' , 'joy' , 'love');`
	q3 := `UPDATE Sample_table SET label = 'bad' WHERE label in ('anger' , 'fear' , 'hate' , 'sadness');`

	if _, err := db.Exec(q1); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q2); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q3); err != nil {
		log.Fatal(err)
	}

	q4 := `UPDATE Usage_table SET label = 'neutral' WHERE label in ('empty' , 'relief' , 'surprise' , 'worry' , 'boredom');`
	q5 := `UPDATE Usage_table SET label = 'good' WHERE label in ('enthusiasm' , 'fun' , 'happiness' , 'joy' , 'love');`
	q6 := `UPDATE Usage_table SET label = 'bad' WHERE label in ('anger' , 'fear' , 'hate' , 'sadness');`

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
func ProcessText(text string) []string {
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
}
