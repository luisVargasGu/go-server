package main

import (
	"net/http"
	"user/server/db"
	"user/server/handlers"
	"github.com/rs/cors"
)

func main() {
    db := db.DbConnect()

    mux := http.NewServeMux()

    mux.HandleFunc("/image", handlers.ImageHandler)

    authHandlerWrapper := func(w http.ResponseWriter, r *http.Request) {
        handlers.AuthHandler(db, w, r)
    }
    
    mux.HandleFunc("/auth", authHandlerWrapper)

    handler := cors.Default().Handler(mux)
    http.ListenAndServe(":8080", handler)
}
