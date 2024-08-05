package middleware

import (
	"log"
	"net/http"
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

func Logger(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		log.Printf("Req: %s %s %s", FormatTracing(r), r.Method, r.URL.Path)
		handler.ServeHTTP(rec, r)
		log.Printf("Res: %s %s %d %v", FormatTracing(r), r.Method, rec.status, time.Since(start))
	}
}
