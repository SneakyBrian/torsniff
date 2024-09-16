package main

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

var (
	db *sql.DB
)

func startIndex() {

	var err error

	db, err = sql.Open("sqlite", "torsniff.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS torrents (
		infohashHex TEXT PRIMARY KEY,
		name TEXT,
		length INTEGER,
		files TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SQLite database initialized")
}
