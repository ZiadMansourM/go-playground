package middleware

import (
	"log"
	"net/http"
	"time"
)

func Timer(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		log.Printf("Total time: %s %v", FormatTracing(r), time.Since(start))
	}
}
