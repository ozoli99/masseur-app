package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var database *sql.DB

func initDatabase() error {
	var err error
	database, err = sql.Open("sqlite", "file:app.db?cache=shared&mode=rwc")
	if err != nil {
		return fmt.Errorf("could not open database: %v", err)
	}

	createTableStmt := `
		CREATE TABLE IF NOT EXISTS appointments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer_name TEXT,
			time DATETIME,
			duration INTEGER,
			notes TEXT
		);
	`

	_, err = database.Exec(createTableStmt)
	if err != nil {
		return fmt.Errorf("could not create table: %v", err)
	}

	return nil
}