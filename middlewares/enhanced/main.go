package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ZiadMansourM/middleware/middleware"
	"github.com/ZiadMansourM/middleware/utils"
)

func UsersMeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from the request context
	user := middleware.AuthenticatedUser(r)
	if user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Do stuff with that user...
	utils.WriteJson(w, http.StatusOK, user)
}

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	router := http.NewServeMux()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJson(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Handle("/users/me", middleware.NewEnsureAuth(http.HandlerFunc(UsersMeHandler)))

	v1 := http.NewServeMux()
	v1.Handle("/api/v1/", http.StripPrefix("/api/v1", router))

	const addr = "127.0.0.1:8080"

	middlewares := []middleware.Middleware{
		middleware.Timer,
		middleware.Logger,
		middleware.Tracer,
	}

	server := http.Server{
		Addr:    addr,
		Handler: middleware.ChainMiddlewares(v1, middlewares...),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", addr, err)
		}
	}()
	log.Printf("Server Listening on %s\n", addr)

	<-done
	fmt.Println("")
	log.Println("Gracefully shutting down server...")

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not shutdown server: %v\n", err)
	}
	log.Println("Server Exited Properly")
}
