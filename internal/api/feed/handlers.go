package feed

import (
	"context"
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/contextkeys"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

type feedService interface {
	List(ctx context.Context, user domain.User) ([]domain.Post, error)
	Post(ctx context.Context, user domain.User, post domain.Post) error
}

func list(f feedService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value(contextkeys.UserKey).(*domain.User)
		if !ok {
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: api.ErrorUnauthorized}
		}

		posts, err := f.List(r.Context(), *user)

		return nil
	}
}

func post(f feedService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value(contextkeys.UserKey).(*domain.User)
		if !ok {
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: api.ErrorUnauthorized}
		}

		err := f.Post(r.Context(), *user, post)
		if err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "failed to post", Err: err}
		}

		return nil
	}
}
