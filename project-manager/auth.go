package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	return string(hash), nil
}

func CreateJWT(userID int64, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(int(userID)),
		"expiresAt": time.Now().Add(time.Hour * 24 * 120).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}

func WithJWTAuth(handlerFunc http.HandlerFunc, store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read JWT from header
		tokenString := r.Header.Get("Authorization")

		// validate token
		// [1]: Check token prefixed with "Bearer "
		// [2]: parse token
		if !strings.HasPrefix(tokenString, "Bearer ") {
			WriteJson(w, http.StatusUnauthorized, ErrorResponse{
				Error: "Unauthorized",
			})
			return
		}

		// strip "Bearer " from token
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := validateToken(tokenString)
		if err != nil {
			WriteJson(w, http.StatusUnauthorized, ErrorResponse{
				Error: "Unauthorized: " + err.Error(),
			})
			return
		}

		if !token.Valid {
			WriteJson(w, http.StatusUnauthorized, ErrorResponse{
				Error: "Unauthorized: invalid token",
			})
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		userID := claims["userID"].(string)

		log.Printf("User ID: %s\n", userID)

		_, err = store.GetUserByID(userID)
		if err != nil {
			WriteJson(w, http.StatusUnauthorized, ErrorResponse{
				Error: "Unauthorized: invalid user. " + err.Error(),
			})
			return
		}
		// call handlerFunc
		handlerFunc(w, r)
	}
}

func validateToken(token string) (*jwt.Token, error) {
	// get secret key
	secret := Envs.JWTSecret
	// parse token
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secret), nil
	})
}
