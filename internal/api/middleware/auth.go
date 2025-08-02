package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/contextkeys"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

type userAccountVerificationService interface {
	ValidateUser(ctx context.Context, token domain.Token) (*domain.User, error)
}

func AuthMiddleware(u userAccountVerificationService) api.MiddlewareFunc {
	return func(method, pattern string, h api.Handler) api.Handler {
		return api.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				return h.ServeHTTP(w, r)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return &api.Error{Code: http.StatusUnauthorized, Message: "Invalid authorization format", Err: api.ErrorUnauthorized}
			}

			token := parts[1]

			user, err := u.ValidateUser(r.Context(), domain.Token(token))
			if err != nil {
				return &api.Error{Code: http.StatusUnauthorized, Message: "Unauthorized", Err: api.ErrorUnauthorized}
			}

			ctx := context.WithValue(r.Context(), contextkeys.UserKey, user)

			return h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
