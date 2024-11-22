package main

import "database/sql"

type Database struct {
	path string
	db   *sql.DB
	rows *sql.Rows
}

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

func CreateCon(path string) (Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return Database{}, err
	}
	var database = Database{path: path, db: db, rows: nil}
	return database, nil
}

func CloseCon(database Database) error {
	var err error
	if database.db != nil {
		err = database.db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
