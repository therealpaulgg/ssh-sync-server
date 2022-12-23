package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Log is a middleware that logs the request and response
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(start)).
			Msg("Request")
	})
}
