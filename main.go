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

	// Connect to database using configuration
	db.Db = db.DbConnect(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	defer db.Db.Close()

	server := api.NewAPIServer(fmt.Sprintf(":%s", cfg.Port), db.Db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
