package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

type loginService interface {
	Login(ctx context.Context, username, password string) (domain.User, error)
}

type userSignInService interface {
	SignIn(ctx context.Context, userConnection domain.UserConnection) (domain.User, error)
}

type userRegisterationService interface {
	Register(ctx context.Context, user domain.User, accountId, password string) error
}

func loginHandler(l loginService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		username, pass, ok := r.BasicAuth()
		if !ok {
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: api.ErrorUnauthorized}
		}

		user, err := l.Login(r.Context(), username, pass)

		switch {
		case errors.Is(err, api.ErrorUnauthorized):
			return &api.Error{Code: http.StatusUnauthorized, Message: "user not authorised", Err: err}
		case err != nil:
			return &api.Error{Code: http.StatusInternalServerError, Message: "user login failed", Err: err}
		}

		api.RespondJSON(w, User{Id: user.Id, DisplayName: user.DisplayName}, http.StatusOK)
		return nil
	}
}

func registerUserHandler(ur userRegisterationService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var newUser RegisterUser
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			return &api.Error{Code: http.StatusBadRequest, Message: "json unmarshal failed", Err: err}
		}

		err := ur.Register(r.Context(), domain.User{DisplayName: newUser.DisplayName},
			newUser.AccountId, newUser.Password)
		if err != nil {
			return &api.Error{Code: http.StatusInternalServerError, Message: "user registeration failed", Err: err}
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		fmt.Fprintf(w, "Pending email verification.")
		return nil
	}
}

func verifyEmailHandler(serverConfig config.ServerConfig, ur userRegisterationService) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, serverConfig.GetDomain()+"/EmailVerificationFailed.html", http.StatusTemporaryRedirect)
		http.Redirect(w, r, serverConfig.GetDomain()+"/EmailVerificationSuccess.html", http.StatusTemporaryRedirect)
		return nil
	}
}
