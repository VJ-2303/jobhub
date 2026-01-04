package main

import (
	"net/http"
	"time"
)

func (app *application) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		app.logger.Info("Request received", "method", r.Method, "uri", r.URL.Path, "duration", duration)
	})
}
