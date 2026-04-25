package api

import (
	"errors"
	"log"
	"net/http"
	"time"
)

func WithLog(logger *log.Logger) MiddlewareFunc {
	return func(method, pattern string, h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			err := h.ServeHTTP(w, r)

			fields := map[string]any{
				"method":   method,
				"pattern":  pattern,
				"url":      r.RequestURI,
				"duration": time.Since(start),
			}

			if err != nil {
				logger.Printf("request failed:\n%v", fields)
				logger.Printf("error: %v", err)
				return err
			}

			logger.Printf("request succeeded:\n%v", fields)
			return err
		})
	}
}

func WithHTTPErrStatus(method, pattern string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		// Limit request body size to 1MB by default
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

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
