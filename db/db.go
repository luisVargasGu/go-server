package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func DbConnect() *sql.DB {
	const (
		host     = "localhost"
		port     = 5432
		user     = "admin"
		password = "password"
		dbname   = "chat_app"
	)

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
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
