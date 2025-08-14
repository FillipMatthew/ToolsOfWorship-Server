package fellowships

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/contextkeys"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

type fellowshipService interface {
	List(ctx context.Context, user domain.User) ([]domain.Fellowship, error)
}

func list(f fellowshipService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := r.Context().Value(contextkeys.UserKey).(*domain.User)
		if !ok {
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: api.ErrorUnauthorized}
		}

		fellowships, err := f.List(r.Context(), *user)
		if err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "could not fetch fellowships", Err: err}
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(fellowships); err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "failed to encode fellowships", Err: err}
		}

		return nil
	}
}
