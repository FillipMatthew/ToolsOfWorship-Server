package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// rateLimiter tracks the request rate for each IP address.
// It uses a simple in-memory map keyed by IP address with a list of request timestamps.
type rateLimiter struct {
	ips map[string][]time.Time
	mu  sync.Mutex
}

// limit checks if an IP has exceeded the rate limit and either allows or blocks the request.
// It removes timestamps older than the time window before checking the count.
func (l *rateLimiter) limit(next Handler, limit int, window time.Duration) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		ip := r.RemoteAddr // Note: Simplistic approach, doesn't handle X-Forwarded-For headers

		l.mu.Lock()
		defer l.mu.Unlock()

		now := time.Now()

		// Initialize request list for this IP if it doesn't exist
		if _, ok := l.ips[ip]; !ok {
			l.ips[ip] = []time.Time{}
		}

		// Remove requests outside the current window
		var validRequests []time.Time
		for _, timestamp := range l.ips[ip] {
			if now.Sub(timestamp) < window {
				validRequests = append(validRequests, timestamp)
			}
		}

		// Check if limit is exceeded
		if len(validRequests) >= limit {
			return &Error{
				Code:    http.StatusTooManyRequests,
				Message: "Too many requests",
				Err:     fmt.Errorf("rate limit exceeded for IP %s", ip),
			}
		}

		// Add current request timestamp
		validRequests = append(validRequests, now)
		l.ips[ip] = validRequests

		return next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware creates a rate limiting middleware that restricts requests from a single IP
// to a maximum of `limit` requests within the specified `window` duration.
//
// Example: RateLimitMiddleware(5, time.Minute) allows 5 requests per minute per IP
func RateLimitMiddleware(limit int, window time.Duration) MiddlewareFunc {
	limiter := &rateLimiter{
		ips: make(map[string][]time.Time),
	}

	return func(method, pattern string, h Handler) Handler {
		return limiter.limit(h, limit, window)
	}
}
