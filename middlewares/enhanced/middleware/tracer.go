package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// Return trace ID and span ID in the following format "[" + traceID + " : " + spanID + "] "
func FormatTracing(r *http.Request) string {
	traceID := r.Header.Get("X-Trace-ID")
	spanID := r.Header.Get("X-Span-ID")
	return "[" + traceID + " : " + spanID + "] "
}

func Tracer(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.NewString()
		r.Header.Set("X-Trace-ID", traceID)
		spanID := uuid.NewString()
		r.Header.Set("X-Span-ID", spanID)
		handler.ServeHTTP(w, r)
	}
}
