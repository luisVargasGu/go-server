package auth

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"time"
	"user/server/types"
)

var jwtSecret = []byte("my_secret_key")

type contextKey string

const UserKey contextKey = "userID"
const UserExpirationKey contextKey = "exp"

func GenerateJWTToken(username string) (string, error) {

	// TODO: Add better expiration time
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

        log.Println("Token:", token)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

        log.Println("TokenString:", tokenString)
	return tokenString, nil
}

func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := AuthenticateRequest(r, store)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		handlerFunc(w, r)
	}
}

func AuthenticateRequest(r *http.Request, s types.UserStore) (*types.User, error) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		log.Println("Error getting JWT token from cookie: ", err)
		return nil, err // No JWT token found in the cookie
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		log.Println("Error parsing JWT token: ", err)
		return nil, err // Failed to parse JWT token
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		expirationTime := claims["exp"].(int64)

		userID, err := strconv.Atoi(username)
		if err != nil {
			log.Println("Error converting username to int: ", err)
			return nil, err
		}

		user, err := s.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user by ID: ", err)
			return nil, err
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, user)
		ctx = context.WithValue(ctx, UserExpirationKey, time.Unix(expirationTime, 0))
		r = r.WithContext(ctx)

		return user, nil
	}

	return nil, jwt.ValidationError{Inner: errors.New("invalid token"), Errors: jwt.ValidationErrorClaimsInvalid}
}
