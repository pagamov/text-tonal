package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	host     = "localhost"
	port     = 5432        // Default PostgreSQL port
	user     = "pagamov"   // Your PostgreSQL username
	password = "multipass" // Your PostgreSQL password
	dbname   = "database"  // Your database name
)

func main() {
	// setup db from backup

	sqliteDB, err := sql.Open("sqlite3", "./main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteDB.Close()

	// Test the connection
	err = sqliteDB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to the database! [1]")

	// Connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open the database connection
	postgreSQL, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer postgreSQL.Close()

	// Test the connection
	err = postgreSQL.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database! [2]")

	deleteAllContent(postgreSQL)
	fmt.Println("Successfully DROP PosgreSQL")

	// TransferLogData(sqliteDB, postgreSQL)
	// log.Println("TransferLogData completed")
	// TransferSampleData(sqliteDB, postgreSQL)
	// log.Println("TransferSampleData completed")
	// TransferUsageData(sqliteDB, postgreSQL)
	// log.Println("TransferUsageData completed")

}

func deleteAllContent(pgDB *sql.DB) {
	pgDB.Exec(`DELETE FROM log_table;`)
	pgDB.Exec(`DELETE FROM sample_table;`)
	pgDB.Exec(`DELETE FROM usage_table;`)
}

func TransferLogData(sqliteDB *sql.DB, pgDB *sql.DB) {
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

func TransferSampleData(sqliteDB *sql.DB, pgDB *sql.DB) {
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

func TransferUsageData(sqliteDB *sql.DB, pgDB *sql.DB) {
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
