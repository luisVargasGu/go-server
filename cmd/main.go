package main

import (
	"user/server/db"
        "user/server/services/hub"
	"user/server/services/image"
	"user/server/services/utils"
)

func main() {
	db.Db = db.DbConnect()
	hub := hub.HubInitialize()
	go hub.Run()


}
