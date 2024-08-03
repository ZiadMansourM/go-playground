package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type UserService struct {
	store Store
}

var errEmailRequired = errors.New("email is required")
var errPasswordRequired = errors.New("password is required")

func NewUserService(store Store) *UserService {
	return &UserService{
		store: store,
	}
}

func (s *UserService) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /users/register", s.handleUserRegistration)
}

func (s *UserService) handleUserRegistration(w http.ResponseWriter, r *http.Request) {
	// get payload: email and password
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, ErrorResponse{
			Error: "Error reading Request Body: " + err.Error(),
		})
		return
	}

	defer r.Body.Close()

	var payload *User
	err = json.Unmarshal(body, &payload)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid Request Payload: " + err.Error(),
		})
		return
	}

	// validate payload
	if err := payload.validate(); err != nil {
		WriteJson(w, http.StatusBadRequest, ErrorResponse{
			Error: "Invalid Request Payload: " + err.Error(),
		})
		return
	}

	hashedPassword, err := HashPassword(payload.Password)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Error hashing password: " + err.Error(),
		})
		return
	}

	payload.Password = hashedPassword

	// create user
	user, err := s.store.CreateUser(payload)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Error creating user: " + err.Error(),
		})
		return
	}

	// Create a token
	_, err = createAndSetAuthCookie(w, user.ID)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, ErrorResponse{
			Error: "Error creating token: " + err.Error(),
		})
		return
	}

	// return user
	WriteJson(w, http.StatusCreated, user)
}

func (u *User) validate() error {
	if u.Email == "" {
		return errEmailRequired
	}

	if u.Password == "" {
		return errPasswordRequired
	}

	return nil
}

func createAndSetAuthCookie(w http.ResponseWriter, userID int64) (string, error) {
	secret := []byte(Envs.JWTSecret)
	token, err := CreateJWT(userID, secret)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		HttpOnly: true,
	})

	return token, nil
}
