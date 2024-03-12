package main

import (
	"net/http"
	"user/server/db"
	"user/server/handlers"
)

func main() {
    db.Db = db.DbConnect()

    mux := http.NewServeMux()

    mux.HandleFunc("/image", handlers.ImageHandler)
    
    mux.HandleFunc("/auth", handlers.CorsHandler(handlers.AuthHandler))

    mux.HandleFunc("/register", handlers.CorsHandler(handlers.RegisterHandler))

    http.ListenAndServe(":8080", mux)
}
