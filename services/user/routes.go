package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"user/server/services/auth"
	"user/server/services/image"
	"user/server/services/utils"
	"user/server/types"

	"github.com/gorilla/mux"
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

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	user, err := decodeRegisterRequestBody(r)
	if err != nil {
		handleError(w, "Invalid request payload", http.StatusBadRequest, err)
		return
	}

	profilePictureData, err := validateAndDecodeImage(user.Avatar)
	if err != nil {
		handleError(w, "Invalid base64 image", http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := hashUserPassword(user.Password)
	if err != nil {
		handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	userID, err := createUserInStore(h.store, user.Email, hashedPassword, profilePictureData)
	if err != nil {
		handleError(w, "Failed to register user, please try again.", http.StatusInternalServerError, err)
		return
	}

	token, err := generateJWTToken(userID)
	if err != nil {
		handleError(w, "Error generating JWT token", http.StatusInternalServerError, err)
		return
	}

	setJWTCookie(w, origin, token)

	prepareResponse(w, userID, user.Avatar, "User registered successfully", http.StatusCreated)
}

func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	creds, err := decodeAuthRequestBody(r)
	if err != nil {
		handleError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	user, err := getUserByEmail(h.store, creds.Email)
	if err != nil {
		handleError(w, "Error getting user by email", http.StatusUnauthorized, err)
		return
	}

	if !comparePasswords(user.Password, []byte(creds.Password)) {
		handleError(w, "Invalid credentials", http.StatusUnauthorized, nil)
		return
	}

	token, err := generateJWTToken(int(user.ID))
	if err != nil {
		handleError(w, "Error generating JWT token", http.StatusInternalServerError, err)
		return
	}

	setJWTCookie(w, origin, token)

	prepareResponse(w, int(user.ID), image.EncodeB64Image(user.Avatar), "Authentication successful", http.StatusOK)
}

func decodeRegisterRequestBody(r *http.Request) (*types.RegisterUserPayload, error) {
	user := &types.RegisterUserPayload{}
	err := json.NewDecoder(r.Body).Decode(user)
	return user, err
}

func decodeAuthRequestBody(r *http.Request) (*types.LoginUserPayload, error) {
	creds := &types.LoginUserPayload{}
	err := json.NewDecoder(r.Body).Decode(creds)
	return creds, err
}

func validateAndDecodeImage(base64Image string) ([]byte, error) {
	return image.DecodeB64Image(base64Image)
}

func hashUserPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

func createUserInStore(store types.UserStore, email, hashedPassword string, profilePictureData []byte) (int, error) {
	user := types.User{
		Username: email,
		Password: hashedPassword,
		Avatar:   profilePictureData,
	}
	return store.CreateUser(user)
}

func getUserByEmail(store types.UserStore, email string) (*types.User, error) {
	return store.GetUserByEmail(email)
}

func comparePasswords(hashedPassword string, plainPassword []byte) bool {
	return auth.ComparePasswords(hashedPassword, plainPassword)
}

func generateJWTToken(userID int) (string, error) {
	return auth.GenerateJWTToken(strconv.Itoa(userID))
}

func setJWTCookie(w http.ResponseWriter, origin, token string) {
	parts := strings.Split(origin, "://")
	domain := "localhost"
	if len(parts) > 1 {
		domainParts := strings.Split(parts[1], ":")
		domain = domainParts[0]
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
		Domain:  domain,
		Path:    "/",
	})
}

func prepareResponse(w http.ResponseWriter, userID int, avatar string, message string, statusCode int) {
	response := types.LoginResponse{
		Success: true,
		Message: message,
		UserID:  userID,
		Avatar:  avatar,
	}

	utils.SendJSONResponse(w, statusCode, response)
}

func handleError(w http.ResponseWriter, message string, statusCode int, err error) {
	if err != nil {
		log.Println(message, err)
	} else {
		log.Println(message)
	}
	http.Error(w, message, statusCode)
}
