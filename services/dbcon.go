package services

import (
	"database/sql"
	"log"
)

var db *sql.DB

func getMySQLDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@(127.0.0.1)/caps?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	return db
}
