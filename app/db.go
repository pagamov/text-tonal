package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

func addLog(text string, label string, info string) {

	currentTime := time.Now()

	// Format the current time to dd.mm.yyyy
	formattedDate := currentTime.Format("02.01.2006")

	// Print the formatted date
	// fmt.Println("Current date in dd.mm.yyyy format:", formattedDate)

	connStr := "user=pagamov password=multipass dbname=database host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	request := `SELECT id FROM log_table
		ORDER BY id DESC
		LIMIT 1;`
	rows, err := db.Query(request)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	var id int
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
	}
	id = id + 1

	_, err = db.Exec("INSERT INTO log_table (id, date, text, label, info) VALUES ($1, $2, $3, $4, $5)",
		id, formattedDate, text, label, info)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Added to log_table", formattedDate, text, label, info)
	}
}

func getLog(date_start string, date_end string) []Statistics {
	connStr := "user=pagamov password=multipass dbname=database host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	start, _ := time.Parse("02.01.2006", date_start)

	date_start = string(start.Format("2006.01.02"))

	end, _ := time.Parse("02.01.2006", date_end)
	date_end = string(end.Format("2006.01.02"))

	fmt.Println(date_start, date_end)

	request := `SELECT date, text, label, info
			FROM log_table
			WHERE date BETWEEN $1 
            AND $2;`

	rows, err := db.Query(request, date_start, date_end)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	var statistics []Statistics
	for rows.Next() {
		var stat Statistics
		var analyz Analyz

		var analyzB string

		if err := rows.Scan(&stat.Date, &stat.Text, &stat.Label, &analyzB); err != nil {
			log.Fatal(err)
		}
		fmt.Println(analyzB)

		json.Unmarshal([]byte(analyzB), &analyz)
		// stat.Words = analyz.Words
		statistics = append(statistics, stat)

	}

	return statistics
}
