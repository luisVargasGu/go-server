package handlers

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"net/http"
        "errors"
	"user/server/db"
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

	credCheck := compareHashedPasswords(creds.Password)
	if !credCheck {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	userID, err := authenticateUser(creds.Username, creds.Password)
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

func authenticateUser(username, password string) (int, error) {
	var userID int

	hashedPassword := hashPassword(password)
	err := db.Db.QueryRow("SELECT id FROM Users WHERE username = $1 AND password_hash = $2", username, hashedPassword).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, err // User not found
	} else if err != nil {
		// database error
		return 0, err
	}

	return userID, nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user Credentials
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	hashedPassword := hashPassword(user.Password)
	// Store user information in your database
	_, err = registerUser(user.Username, hashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func hashPassword(password string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		return []byte(hashedPassword)
	}
	return nil
}

// false is passwords don't match
// true is passwords match
func compareHashedPasswords(password string) bool {
	hashedPassword := hashPassword(password)
	if hashedPassword == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return false
	}
	return true
}

func registerUser(username string, password []byte) (sql.Result, error) {
	// Check if the user already exists
	var count int
	err := db.Db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("user already exists")
	}

	res, err := db.Db.Exec("INSERT INTO Users VALUES ($1, $2)", username, password)
	if err != nil {
		return nil, err
	}
	return res, nil
}
