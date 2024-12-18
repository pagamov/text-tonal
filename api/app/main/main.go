package main

import (
	"api/db"
	"api/model"
	"api/router"

	_ "github.com/mattn/go-sqlite3"

	// _ "modernc.org/sqlite"
	_ "github.com/jmoiron/sqlx" // Load .env file
	_ "github.com/lib/pq"
)

var (
	TextModel      model.Model
	api            router.Router
	databaseSqlite db.DatabaseSQLite
)

func main() {

	databaseSqlite = *db.CreateDatabaseSQLite("../../../db/main.db")
	databaseSqlite.Init()
	databaseSqlite.ReplaceLabels()
	databaseSqlite.PrintLabels()

	TextModel.Init(databaseSqlite)

	testData := TextModel.LearnNew(databaseSqlite, 0.8, false)

	TextModel.Test(testData)

	// api.Init()
	// api.AddMethod()
	// api.Start("8080")
}
