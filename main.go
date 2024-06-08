package main

import (
	"log"
	"user/server/cmd/api"
	"user/server/db"
)

func main() {
	db.Db = db.DbConnect()
	defer db.Db.Close()

	server := api.NewAPIServer(":8080", db.Db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
