package main

import (
	"net/http"
	"user/server/db"
	"user/server/handlers"
	"github.com/rs/cors"
)

func main() {
    db.Db = db.DbConnect()

    mux := http.NewServeMux()

    mux.HandleFunc("/image", handlers.ImageHandler)
    
    mux.HandleFunc("/auth", handlers.AuthHandler)

    mux.HandleFunc("/register", handlers.RegisterHandler)

    handler := cors.Default().Handler(mux)
    http.ListenAndServe(":8080", handler)
}
