package appserver

import (
	"encoding/json"
	"golang-share-files/internal/apperror"
	"log"
	"net/http"
	"strings"
	"time"
)

func withToken(expToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		token := parts[1]

		if token != expToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(apperror.ErrorResponse{
				Code: "BAD_REQUEST",
				Msg:  "Token tidak ada atau tidak valid",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}

func withMaxBody(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		wrappedWriter := &apperror.StatusResponseWriter{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(startTime)

		log.Printf("[%s] - %s %s - %d - %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrappedWriter.Status,
			duration,
		)
	})
}
