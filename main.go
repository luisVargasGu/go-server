package main

import (
    "net/http"
    "github.com/rs/cors"
)

func main() {
    DbConnect()
    mux := http.NewServeMux()

    mux.HandleFunc("/image", ImageHandler)
    mux.HandleFunc("/auth", AuthHandler)

    handler := cors.Default().Handler(mux)
    http.ListenAndServe(":8080", handler)
}
