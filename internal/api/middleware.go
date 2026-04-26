package api

import (
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to all HTTP responses.
// This includes HSTS, X-Content-Type-Options, X-Frame-Options, and others.
func SecurityHeadersMiddleware(method, pattern string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Strict-Transport-Security tells browsers to always use HTTPS for the next 2 years
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking by disallowing framing
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS filtering in legacy browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Restrict resource loading to same origin by default
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		return h.ServeHTTP(w, r)
	})
}

// CORSMiddleware returns a MiddlewareFunc that handles Cross-Origin Resource Sharing (CORS).
//
// Allowed origin behaviour:
//   - If allowedOrigins is empty, all origins are permitted and the response carries
//     "Access-Control-Allow-Origin: *".
//   - Otherwise the request's Origin header is checked against the list. If it matches,
//     that specific origin is echoed back and a "Vary: Origin" header is added so that
//     caches keep per-origin copies of the response.
//   - If the origin does not match, no CORS headers are added and the request continues
//     normally (the browser will enforce the same-origin policy itself).
func CORSMiddleware(allowedOrigins []string) MiddlewareFunc {
	// Build a set for O(1) look-ups.
	allowAll := len(allowedOrigins) == 0
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(method, pattern string, h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			origin := r.Header.Get("Origin")

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if _, ok := allowed[origin]; ok {
					// Echo the matched origin so cookies/credentials work correctly.
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Tell caches that the response differs per origin.
					w.Header().Add("Vary", "Origin")
				}
				// Unknown origin: omit CORS headers; the browser enforces the policy.
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Short-circuit OPTIONS preflight requests immediately.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return nil
			}

			return h.ServeHTTP(w, r)
		})
	}
}
