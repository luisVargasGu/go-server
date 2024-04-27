package user

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"user/server/services/auth"
	"user/server/services/utils"
	"user/server/types"
)

type Handler struct {
	store types.UserStore
}

func NewHandler(store types.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/auth", utils.CorsHandler(h.AuthHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/register", utils.CorsHandler(h.RegisterHandler)).Methods("POST", "OPTIONS")
}

func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	var creds types.LoginUserPayload
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Error decoding request body: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByEmail(creds.Email)
	if err != nil {
		log.Println("Error getting user by email: ", err)
		http.Error(w, "Error getting user by email", http.StatusUnauthorized)
		return
	}

	if !auth.ComparePasswords(user.Password, []byte(creds.Password)) {
		log.Println("Invalid credentials")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWTToken(strconv.Itoa(int(user.ID)))
	if err != nil {
		log.Println("Error generating JWT token: ", err)
		http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
		return
	}

	parts := strings.Split(origin, "://")
	domain := "localhost"
	if len(parts) > 1 {
		domainParts := strings.Split(parts[1], ":")
		domain = domainParts[0]
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
		Domain:  domain,
		Path:    "/",
	})

	response := types.LoginResponse{
		Success: true,
		Message: "Authentication successful",
		UserID:  user.ID,
	}

	utils.SendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	var user types.RegisterUserPayload
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("Error decoding request body: ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	log.Println("User:", user)
	userID, err := h.store.CreateUser(types.User{Username: user.Email, Password: hashedPassword})
	if err != nil {
		log.Println("Error registering user: ", err)
		http.Error(w, "Failed to register user, please try again.", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateJWTToken(strconv.Itoa(userID))
	if err != nil {
		log.Println("Error generating JWT token: ", err)
		http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
		return
	}

	parts := strings.Split(origin, "://")
	domain := "localhost"
	if len(parts) > 1 {
		domainParts := strings.Split(parts[1], ":")
		domain = domainParts[0]
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
		Domain:  domain,
		Path:    "/",
	})

	response := types.LoginResponse{
		Success: true,
		Message: "User registered successfully",
		UserID:  userID,
	}

	utils.SendJSONResponse(w, http.StatusCreated, response)
}
