package handlers

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"
	"user/server/db"
)

var jwtSecret = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	UserID  int64  `json:"user_id,omitempty"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if r.Method == http.MethodOptions {
		// Handle preflight request
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Error decoding request body: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := authenticateUser(creds.Username, creds.Password)
	if err != nil {
		log.Println("Error authenticating user: ", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWTToken(strconv.Itoa(int(userID)))
	if err != nil {
		log.Println("Error generating JWT token: ", err)
		http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour), // Example: Set expiration time for 24 hours
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

	rows, err := db.Db.Query("SELECT password, user_id FROM \"Users\" WHERE username = $1", username)
	if err != nil {
		log.Println("Error querying database: ", err)
		return 0, err // User not found
	}

	defer rows.Close()

	for rows.Next() {
		var dbPassword string
		err = rows.Scan(&dbPassword, &userID)
		if err != nil {
			log.Println("Error scanning rows: ", err)
			return 0, err
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err == nil {
			// Authentication successful, return the user ID
			return userID, nil
		}
	}
	log.Println("Invalid credentials")
	return 0, errors.New("invalid credentials")
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if r.Method == http.MethodOptions {
		// Handle preflight request
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user Credentials
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("Error decoding request body: ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(user.Password)

	userID, err := registerUser(user.Username, hashedPassword)
	if err != nil {
		log.Println("Error registering user: ", err)
		http.Error(w, "Failed to register user, please try again.", http.StatusInternalServerError)
		return
	}

	token, err := generateJWTToken(strconv.Itoa(int(userID)))
	if err != nil {
		log.Println("Error generating JWT token: ", err)
		http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
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
		return hashedPassword
	}
	return nil
}

// TODO: add same username support
func registerUser(username string, password []byte) (int64, error) {
	// Check if the user already exists

	var count int
	err := db.Db.QueryRow("SELECT COUNT(*) FROM \"Users\" WHERE username = $1", username).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("user already exists")
	}

	var userID int64
	err = db.Db.QueryRow("INSERT INTO \"Users\" (username, password) VALUES ($1, $2) RETURNING user_id", username, password).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
