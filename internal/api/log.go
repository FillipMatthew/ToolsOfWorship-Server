package api

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// WithLog middleware logs request details
func WithLog(logger *slog.Logger) MiddlewareFunc {
	return func(method, pattern string, h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			err := h.ServeHTTP(w, r)

			if err != nil {
				logger.Error("request failed",
					slog.String("method", method),
					slog.String("pattern", pattern),
					slog.String("url", r.RequestURI),
					slog.Duration("duration", time.Since(start)),
					slog.Any("error", err),
				)
				return err
			}

			logger.Info("request succeeded",
				slog.String("method", method),
				slog.String("pattern", pattern),
				slog.String("url", r.RequestURI),
				slog.Duration("duration", time.Since(start)),
			)
			return err
		})
	}
}

// WithHTTPErrStatus middleware handles HTTP error responses
func WithHTTPErrStatus(method, pattern string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		err := h.ServeHTTP(w, r)
		if err == nil {
			return nil
		}

		var apiErr *Error

		if !errors.As(err, &apiErr) {
			apiErr = &Error{Code: http.StatusInternalServerError, Message: "Something went wrong"}
		} else {
			// Don't leak the underlying error if it's internal
			if apiErr.Code == http.StatusInternalServerError {
				apiErr.Message = "Something went wrong"
			}
		}

		RespondError(w, apiErr)

		return err
	})
}

// WithBodyLimit returns a MiddlewareFunc that limits the request body size for a specific handler.
func WithBodyLimit(maxBytes int64) MiddlewareFunc {
	return func(method, pattern string, h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			return h.ServeHTTP(w, r)
		})
	}
}
