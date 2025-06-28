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
		err := h.ServeHTTP(w, r)
		if err == nil {
			return nil
		}

		var apiErr *Error

		if !errors.As(err, &apiErr) {
			apiErr = &Error{Code: http.StatusInternalServerError, Message: "something went wrong", Err: err}
		}

		RespondError(w, apiErr)

		return err
	})
}
