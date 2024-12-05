package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jbrukh/bayesian"
	_ "github.com/mattn/go-sqlite3"
)

// Database class

type Database struct {
	path string
	db   *sql.DB
	rows *sql.Rows
}

// setup all db tables
func (database Database) init() {
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
}

func (database Database) getTestData() ([]Data, error) {
	var err error
	var res []Data = []Data{}
	var db *sql.DB
	var rows *sql.Rows
	var querry string = `
		SELECT label, text_ru from Sample_table where text_ru is not NULL;
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
func (database Database) getWordsByLabel(label string) ([]string, error) {
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
func (database Database) getUsage(labels []bayesian.Class) ([]Data, error) {
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
func (database Database) getLabels() ([]bayesian.Class, error) {
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

// func CloseCon(database Database) error {
// 	var err error
// 	if database.db != nil {
// 		err = database.db.Close()
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
