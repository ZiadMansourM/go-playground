package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt"
)

type APIServer struct {
	addr  string
	store Store
}

func NewAPIServer(addr string, store Store) *APIServer {
	return &APIServer{
		addr:  addr,
		store: store,
	}
}

func (s *APIServer) Run() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here e.g.:
		// Close database, redis, truncate message queues, etc.
		cancel()
	}()

	router := http.NewServeMux()

	v1 := http.NewServeMux()
	v1.Handle("/api/v1/", http.StripPrefix("/api/v1", router))

	// Health Check
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJson(w, http.StatusOK, map[string]string{"message": "API is healthy"})
	})

	// START Registering Services
	tasksService := NewTasksService(s.store)
	tasksService.RegisterRoutes(router)

	usersService := NewUserService(s.store)
	usersService.RegisterRoutes(router)
	// END Registering Services

	middlewareChain := MiddlewareChain(
		RequestLoggerMiddleware,
		RequireAuthMiddleware,
	)

	server := http.Server{
		Addr:    s.addr,
		Handler: middlewareChain(v1),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", s.addr, err)
		}
	}()
	log.Printf("Server Listening on %s\n", s.addr)

	<-done
	fmt.Println("")
	log.Println("Gracefully shutting down server...")

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not shutdown server: %v\n", err)
	}
	log.Println("Server Exited Properly")
}

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func RequireAuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Exclude user registration route from authentication
		if r.URL.Path == "/api/v1/users/register" {
			next.ServeHTTP(w, r)
			return
		}

		// Read JWT from header
		tokenString := r.Header.Get("Authorization")

		// validate token
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

		// _, err = store.GetUserByID(userID)
		// if err != nil {
		// 	WriteJson(w, http.StatusUnauthorized, ErrorResponse{
		// 		Error: "Unauthorized: invalid user. " + err.Error(),
		// 	})
		// 	return
		// }

		next.ServeHTTP(w, r)
	}
}

type Middleware func(http.Handler) http.HandlerFunc

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next.ServeHTTP
	}
}
