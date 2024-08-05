package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func WriteJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

type Logger struct {
	handler http.Handler
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rec := &statusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
	l.handler.ServeHTTP(rec, r)
	log.Printf("%s %s %d %v", r.Method, r.URL.Path, rec.status, time.Since(start))
}

func NewLogger(handlerToWrap http.Handler) *Logger {
	return &Logger{handlerToWrap}
}

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
		WriteJson(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := http.NewServeMux()
	v1.Handle("/api/v1/", http.StripPrefix("/api/v1", router))

	const addr = "127.0.0.1:8080"

	server := http.Server{
		Addr:    addr,
		Handler: NewLogger(v1),
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
