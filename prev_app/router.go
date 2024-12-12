package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
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

type API struct {
	router *gin.Engine
}

func (api *API) init() {
	api.router = gin.Default()
}

func (api *API) addMethod() {
	api.router.POST("/analyze", analyze)
	api.router.GET("/statistics", statistics)
	api.router.POST("/db/import_from_old", import_from_old)
	api.router.POST("/db/transfer_sqlite_to_posgresql", transfer_sqlite_to_posgresql)
}

func (api *API) start(port string) {
	api.router.Run(fmt.Sprintf(":%s", port))
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

func import_from_old(c *gin.Context) {
	// Connect to the main database
	db, err := sql.Open("sqlite3", "../db/main.db")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
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
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	// Create Sample_table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "Sample_table" (
		"id" INTEGER PRIMARY KEY,
		"text_en" TEXT,
		"text_ru" TEXT DEFAULT NULL,
		"label" TEXT,
		"processed" INTEGER DEFAULT 0
	);`)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	// Clear Sample_table
	_, err = db.Exec(`DELETE FROM "Sample_table";`)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	// Insert data from old databases
	var g errgroup.Group
	for i := 0; i < 10; i++ {
		i := i // capture range variable
		g.Go(func() error {
			conOld, err := sql.Open("sqlite3", filepath.Join("../homework/data/db", fmt.Sprintf("mydatabase_%d.db", i)))
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
				if err := rows.Scan(&id, &label, &textEn, &textRu); err != nil {
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
		c.IndentedJSON(http.StatusBadRequest, err)
		return
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
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	// Clear Usage_table
	_, err = db.Exec(`DELETE FROM "Usage_table";`)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}

	// Select rows from Sample_table
	rows, err := db.Query(`SELECT * FROM "Sample_table" WHERE "text_en" IS NOT NULL;`)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	allUsage := make(map[string]map[string]int)

	for rows.Next() {
		var id int
		var textEn, textRu, label string
		var processed int
		if err := rows.Scan(&id, &textEn, &textRu, &label, &processed); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err)
			return
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
			_, err = db.Exec("INSERT INTO \"Usage_table\" (word, language, label, usage) VALUES (?, ?, ?, ?)", word, "en", label, count)
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, err)
				return
			}
		}
	}

	fmt.Println("Data processing complete.")

	c.IndentedJSON(http.StatusOK, "done")
}

func transfer_sqlite_to_posgresql(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
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
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	defer pgDB.Close()

	err = pgDB.Ping()
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	fmt.Println("Successfully connected to the database!")

	_, err = pgDB.Exec(`DO $$ 
DECLARE 
    r RECORD; 
BEGIN 
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP 
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; 
    END LOOP; 
END $$;`)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	log.Print("old data deleted")

	// Create tables in PostgreSQL
	_, err = pgDB.Exec(LogTableSQL)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	// log.Print("Log done")
	_, err = pgDB.Exec(SampleTableSQL)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	// log.Print("Sample done")
	_, err = pgDB.Exec(UsageTableSQL)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	// log.Print("Usage done")

	// Connect to SQLite
	sqliteDB, err := sql.Open("sqlite3", "../db/main.db")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err)
		return
	}
	defer sqliteDB.Close()

	// Transfer data from Log_table
	transferLogData(sqliteDB, pgDB)

	// log.Print("Log done")

	// Transfer data from Sample_table
	transferSampleData(sqliteDB, pgDB)

	// log.Print("Sample done")

	// Transfer data from Usage_table
	transferUsageData(sqliteDB, pgDB)

	// log.Print("Usage done")

	c.IndentedJSON(http.StatusOK, "Data transfer completed successfully!")
}
