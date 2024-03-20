package main

import (
	"net/http"
	"user/server/db"
	"user/server/handlers"
)

func main() {
    db.Db = db.DbConnect()
    hub := handlers.HubInitialize()
    go hub.Run()

    mux := http.NewServeMux()

    mux.HandleFunc("/image", handlers.ImageHandler)
    
    mux.HandleFunc("/auth", handlers.CorsHandler(handlers.AuthHandler))

    mux.HandleFunc("/register", handlers.CorsHandler(handlers.RegisterHandler))

    mux.HandleFunc("/chat", handlers.ChattingHandler)

    http.ListenAndServe(":8080", mux)
}
