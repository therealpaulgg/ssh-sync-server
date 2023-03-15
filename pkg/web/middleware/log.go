package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func getUserIP(req *http.Request) net.IP {
	var userIP string
	if len(req.Header.Get("X-Forwarded-For")) > 1 {
		userIP = req.Header.Get("X-Forwarded-For")
		return net.ParseIP(userIP)
	} else if len(req.Header.Get("X-Real-IP")) > 1 {
		userIP = req.Header.Get("X-Real-IP")
		return net.ParseIP(userIP)
	} else {
		userIP = req.RemoteAddr
		if strings.Contains(userIP, ":") {
			return net.ParseIP(strings.Split(userIP, ":")[0])
		} else {
			return net.ParseIP(userIP)
		}
	}
}

// Log is a middleware that logs the request and response
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", getUserIP(r).String()).
			Dur("duration", time.Since(start)).
			Msg("Request")
	})
}
