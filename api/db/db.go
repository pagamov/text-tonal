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

const ()

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

func (database *DatabaseSQLite) ReplaceLabels() {
	var db *sql.DB
	var err error
	var path string = database.path

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// q1 := `UPDATE Sample_table SET label = 'neutral' WHERE label in ();`
	q2 := `UPDATE Sample_table SET label = 'good' WHERE label in ('relief', 'enthusiasm' , 'fun' , 'happiness' , 'joy' , 'love', 'surprise');`
	q3 := `UPDATE Sample_table SET label = 'bad' WHERE label in ('anger' , 'fear' , 'hate' , 'sadness', 'worry', 'boredom', 'empty');`

	// if _, err := db.Exec(q1); err != nil {
	// 	log.Fatal(err)
	// }
	if _, err := db.Exec(q2); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(q3); err != nil {
		log.Fatal(err)
	}

	// q4 := `UPDATE Usage_table SET label = 'neutral' WHERE label in ('empty' , 'relief' , 'surprise' , 'worry' , 'boredom');`
	q5 := `UPDATE Usage_table SET label = 'good' WHERE label in ('relief', 'enthusiasm' , 'fun' , 'happiness' , 'joy' , 'love', 'surprise');`
	q6 := `UPDATE Usage_table SET label = 'bad' WHERE label in ('anger' , 'fear' , 'hate' , 'sadness', 'worry', 'boredom', 'empty');`

	// if _, err := db.Exec(q4); err != nil {
	// 	log.Fatal(err)
	// }
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
