package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/contextkeys"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

type feedService interface {
	List(ctx context.Context, user domain.User, limit *int, before *time.Time, after *time.Time) ([]domain.Post, error)
	Post(ctx context.Context, user domain.User, fellowshipId uuid.UUID, circleId uuid.UUID, heading string, article string) error
}

func list(f feedService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value(contextkeys.UserKey).(*domain.User)
		if !ok {
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: api.ErrorUnauthorized}
		}

		var listRequest ListRequest
		if err := json.NewDecoder(r.Body).Decode(&listRequest); err != nil {
			return &api.Error{Code: http.StatusBadRequest, Message: "json unmarshal failed", Err: err}
		}

		posts, err := f.List(r.Context(), *user, listRequest.Limit, listRequest.Before, listRequest.After)
		if err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "could not fetch posts", Err: err}
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(posts); err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "failed to encode posts", Err: err}
		}

		return nil
	}
}

func post(f feedService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value(contextkeys.UserKey).(*domain.User)
		if !ok {
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: api.ErrorUnauthorized}
		}

		var postRequest PostRequest
		if err := json.NewDecoder(r.Body).Decode(&postRequest); err != nil {
			return &api.Error{Code: http.StatusBadRequest, Message: "json unmarshal failed", Err: err}
		}

		err := f.Post(r.Context(), *user, postRequest.FellowshipId, postRequest.CircleId, postRequest.Heading, postRequest.Article)
		if err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "failed to post", Err: err}
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}
