package config

import (
	"os"
)

type Config struct {
	Environment string
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
}

var (
	Development = Config{
		Environment: "development",
		Port:        "8080",
		DBHost:      "localhost",
		DBPort:      "5432",
		DBUser:      "postgres",
		DBPassword:  "password",
		DBName:      "chat_app",
	}

	Production = Config{
		Environment: "production",
		Port:        os.Getenv("PORT"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
	}
)

func GetConfig() Config {
	env := os.Getenv("GO_ENV")
	if env == "production" {
		return Production
	}
	return Development
}
