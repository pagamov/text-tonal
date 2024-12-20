package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Database class

type DatabaseSQLite struct {
	path string
	db   *sql.DB
	rows *sql.Rows
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
