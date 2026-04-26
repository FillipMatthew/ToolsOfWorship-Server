package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// clientIP extracts the real client IP address from a request.
// It checks, in order:
//  1. X-Real-IP header (set by trusted proxies such as Nginx)
//  2. X-Forwarded-For header — takes the leftmost (original client) address
//  3. r.RemoteAddr, stripping the port via net.SplitHostPort
func clientIP(r *http.Request) string {
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be a comma-separated list; the leftmost entry is
		// the original client IP appended by the first proxy that received the
		// request.
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Fall back to the direct connection address, dropping the port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr has no port (unlikely but possible) — use it as-is.
		return r.RemoteAddr
	}
	return host
}

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
		ip := clientIP(r)

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
