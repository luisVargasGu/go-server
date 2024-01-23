package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"os"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.URL.Query().Get("id")
	imagePath := "images/" + imageID + ".jpg"

	file, err := os.Open(imagePath)
	if err != nil {
		http.Error(w, "Image not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error serving image", http.StatusInternalServerError)
		return
	}
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message.omitempty"`
	UserID  int    `json:"user_id.omitempty"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
        //TODO: check creds against database
	if creds.Username == "admin" && creds.Password == "password" {
		response := AuthResponse{
			Success: true,
			Message: "Authentication successful",
			UserID:  1, 
		}
		sendJSONResponse(w, http.StatusOK, response)
	} else {
		response := AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		}
		sendJSONResponse(w, http.StatusUnauthorized, response)
	}
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func DbConnect() {
	const (
		host     = "localhost"
		port     = 5432
		user     = "admin"
		password = "password"
		dbname   = "chat_app"
	)

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
