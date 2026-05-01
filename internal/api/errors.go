package api

import (
	"errors"
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

var (
	ErrorUnauthorized = errors.New("unauthorized")
	//ErrorNotFound     = errors.New("user not found")
)

func MapDomainError(err error) *Error {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail):
		return &Error{Code: http.StatusUnprocessableEntity, ErrorCode: "invalid_email", Message: "invalid email format", Err: err}
	case errors.Is(err, domain.ErrInvalidDisplayName):
		return &Error{Code: http.StatusUnprocessableEntity, ErrorCode: "invalid_display_name", Message: "invalid display name", Err: err}
	case errors.Is(err, domain.ErrPasswordTooShort):
		return &Error{Code: http.StatusUnprocessableEntity, ErrorCode: "password_too_short", Message: "password too short", Err: err}
	case errors.Is(err, domain.ErrEmailInUse):
		return &Error{Code: http.StatusConflict, ErrorCode: "email_in_use", Message: "email already in use", Err: err}
	case errors.Is(err, domain.ErrInvalidCredentials):
		return &Error{Code: http.StatusUnauthorized, ErrorCode: "invalid_credentials", Message: "invalid credentials", Err: err}
	case errors.Is(err, domain.ErrInvalidToken):
		return &Error{Code: http.StatusUnauthorized, ErrorCode: "invalid_token", Message: "invalid token", Err: err}
	case errors.Is(err, domain.ErrUserNotFound):
		return &Error{Code: http.StatusNotFound, ErrorCode: "user_not_found", Message: "user not found", Err: err}
	default:
		return &Error{Code: http.StatusInternalServerError, Message: "internal error", Err: err}
	}
}
