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
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here e.g.:
		// Close database, redis, truncate message queues, etc.
		cancel()
	}()

	router := http.NewServeMux()

	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJson(w, http.StatusOK, map[string]string{"message": "Hello, World!"})
	})

	v1 := http.NewServeMux()
	v1.Handle("/api/v1/", http.StripPrefix("/api/v1", router))

	middlewareChain := MiddlewareChain(
		RequestLoggerMiddleware,
		// RequireAuthMiddleware,
	)

	const addr = "127.0.0.1:8080"

	server := http.Server{
		Addr:    addr,
		Handler: middlewareChain(v1),
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
