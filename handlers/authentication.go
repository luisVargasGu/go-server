package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"user/server/db"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message.omitempty"`
	UserID  int64  `json:"user_id.omitempty"`
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
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWTToken(strconv.Itoa(int(userID)))
	if err != nil {
		http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // Example: Set expiration time for 24 hours
		HttpOnly: true,
		Secure:   true, // Set true if your application is using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	response := AuthResponse{
		Success: true,
		Message: "Authentication successful",
		UserID:  userID,
	}

	sendJSONResponse(w, http.StatusOK, response)
}

func authenticateUser(username, password string) (int64, error) {
	var userID int64

	hashedPassword := hashPassword(password)
	err := db.Db.QueryRow("SELECT user_id FROM Users WHERE username = $1 AND password_hash = $2", username, hashedPassword).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, err // User not found
	} else if err != nil {
		// database error
		return 0, err
	}

	return userID, nil
}

func RegisterHanrler(w http.ResponseWriter, r *http.Request) {
	var user Credentials
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(user.Password)

	userID, err := registerUser(user.Username, hashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	token, err := generateJWTToken(strconv.Itoa(int(userID)))
	if err != nil {
		http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	response := AuthResponse{
		Success: true,
		Message: "User registered successfully",
		UserID:  userID,
	}

	sendJSONResponse(w, http.StatusCreated, response)
}

func generateJWTToken(username string) (string, error) {

	// TODO: Add better expiration time
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
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

func registerUser(username string, password []byte) (int64, error) {
	// Check if the user already exists
	var count int
	err := db.Db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("user already exists")
	}

	var userID int64
	err = db.Db.QueryRow("INSERT INTO Users(username, password) VALUES ($1, $2) RETURNING user_id", username, password).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
