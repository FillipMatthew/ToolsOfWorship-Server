package api

import (
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to all HTTP responses
func SecurityHeadersMiddleware(method, pattern string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// HSTS - enforce HTTPS for future requests
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy - restrict resource loading
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		return h.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS (Cross-Origin Resource Sharing) headers and preflight requests
func CORSMiddleware(method, pattern string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// TODO: Make this configurable based on environment/allowed origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle CORS preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return h.ServeHTTP(w, r)
	})
}
