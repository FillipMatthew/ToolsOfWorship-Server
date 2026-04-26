package api

import (
	"context"
	"net/http"
	"time"
)

// WithTimeout returns a middleware that cancels the request context after the given duration.
// All downstream handlers — including database calls — inherit this deadline, so no single
// request can block resources indefinitely.
//
// When the deadline is exceeded the context is cancelled; handlers that respect context
// cancellation will return early. The HTTP response for a timed-out request is written
// by the WithHTTPErrStatus middleware once the error propagates back up the chain.
func WithTimeout(d time.Duration) MiddlewareFunc {
	return func(method, pattern string, h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()

			return h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
