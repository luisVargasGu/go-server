package handlers

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"net/http"
)

var jwtSecret = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message.omitempty"`
	UserID  int    `json:"user_id.omitempty"`
}

func AuthHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	userID, err := authenticateUser(db, creds.Username, creds.Password)
	if err == nil {
		response := AuthResponse{
			Success: true,
			Message: "Authentication successful",
			UserID:  userID,
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

func authenticateUser(db *sql.DB, username, password string) (int, error) {
	var userID int

	err := db.QueryRow("SELECT id FROM Users WHERE username = $1 AND password_hash = $2", username, password).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, err // User not found
	} else if err != nil {
		// database error
		return 0, err
	}

	return userID, nil
}
