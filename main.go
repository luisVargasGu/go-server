package main

import (
	"log"
	"user/server/cmd/api"
	"user/server/db"
	"user/server/services/hub"
)

func main() {
	db.Db = db.DbConnect()
	hub := hub.HubInitialize()
	go hub.Run()

	server := api.NewAPIServer(":8080", db.Db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
