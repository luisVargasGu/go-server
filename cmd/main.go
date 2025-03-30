package main

import (
	"fmt"
	"log"
	"user/server/cmd/api"
	"user/server/config"
	"user/server/db"
)

func main() {
	cfg := config.GetConfig()
	db.Db = db.DbConnect(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	server := api.NewAPIServer(fmt.Sprintf(":%s", cfg.Port), db.Db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
