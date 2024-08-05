package middleware

import (
	"context"
	"net/http"
)

type EnsureAuth struct {
	handler http.Handler
}

type User struct {
	ID   int
	Name string
}

var authenticatedUserKey = &contextKey{"authenticated-user"}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "middleware context key " + k.name
}

func GetAuthenticatedUser(r *http.Request) (*User, error) {
	// For demonstration purposes, we assume the user is always authenticated.
	// In a real scenario, you would check cookies, tokens, etc.
	return &User{ID: 1, Name: "John Doe"}, nil
}

func (ea *EnsureAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, "please sign-in", http.StatusUnauthorized)
		return
	}

	// Create a new request context containing the authenticated user
	ctxWithUser := context.WithValue(r.Context(), authenticatedUserKey, user)
	// Create a new request using that new context
	rWithUser := r.WithContext(ctxWithUser)
	// Call the real handler, passing the new request
	ea.handler.ServeHTTP(w, rWithUser)
}

func NewEnsureAuth(handlerToWrap http.Handler) *EnsureAuth {
	return &EnsureAuth{handlerToWrap}
}

func AuthenticatedUser(r *http.Request) *User {
	user, _ := r.Context().Value(authenticatedUserKey).(*User)
	return user
}
