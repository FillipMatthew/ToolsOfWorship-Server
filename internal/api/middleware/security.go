package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
)

func SecurityHeadersMiddleware(method, pattern string, h api.Handler) api.Handler {
	return api.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		return h.ServeHTTP(w, r)
	})
}

func CORSMiddleware(method, pattern string, h api.Handler) api.Handler {
	return api.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Ideally make this configurable
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		return h.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	ips map[string][]time.Time
	mu  sync.Mutex
}

func (l *rateLimiter) limit(next api.Handler, limit int, window time.Duration) api.Handler {
	return api.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ip := r.RemoteAddr // Simplistic, doesn't handle X-Forwarded-For

		l.mu.Lock()
		now := time.Now()
		if _, ok := l.ips[ip]; !ok {
			l.ips[ip] = []time.Time{}
		}

		// Cleanup old requests
		var valid []time.Time
		for _, t := range l.ips[ip] {
			if now.Sub(t) < window {
				valid = append(valid, t)
			}
		}

		if len(valid) >= limit {
			l.mu.Unlock()
			return &api.Error{Code: http.StatusTooManyRequests, Message: "Too many requests", Err: fmt.Errorf("rate limit exceeded")}
		}

		valid = append(valid, now)
		l.ips[ip] = valid
		l.mu.Unlock()

		return next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(limit int, window time.Duration) api.MiddlewareFunc {
	limiter := &rateLimiter{ips: make(map[string][]time.Time)}
	return func(method, pattern string, h api.Handler) api.Handler {
		return limiter.limit(h, limit, window)
	}
}
