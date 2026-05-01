package domain

import (
	"regexp"
	"time"
)

const (
	EmailRegexPattern       = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	DisplayNameRegexPattern = `^[a-zA-Z0-9 ]{3,30}$`
	DisplayNameMinLength    = 3
	DisplayNameMaxLength    = 30
	PasswordMinLength       = 8

	// TokenExpiryDuration is how long a user auth/verification token remains valid.
	TokenExpiryDuration = 15 * time.Minute

	// KeyExpiryDuration is how long a signing/encryption key is valid (182 days ≈ 6 months).
	KeyExpiryDuration = 182 * 24 * time.Hour
)

// EmailRegex is the compiled form of EmailRegexPattern, ready for use without re-compilation.
var EmailRegex = regexp.MustCompile(EmailRegexPattern)

// DisplayNameRegex is the compiled form of DisplayNameRegexPattern.
var DisplayNameRegex = regexp.MustCompile(DisplayNameRegexPattern)
