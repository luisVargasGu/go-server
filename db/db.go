package db

import (
	"database/sql"
	"fmt"
	"log"
)

var Db *sql.DB

func DbConnect(host, port, user, password, dbname string) *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database:", err)
	}

	fmt.Println("Successfully connected!")
	return db
}
