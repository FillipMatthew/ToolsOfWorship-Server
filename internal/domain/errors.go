package domain

import "errors"

var (
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailInUse         = errors.New("email already in use")

	// Validation errors
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrPasswordTooShort   = errors.New("password too short")

	// Token errors
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenData = errors.New("invalid token data")

	// User errors
	ErrUserNotFound    = errors.New("user not found")
	ErrUserFetchFailed = errors.New("unable to fetch user")
)
