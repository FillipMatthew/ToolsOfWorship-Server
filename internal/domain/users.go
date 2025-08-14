package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SignInType int32

const (
	None SignInType = iota
	LocalUser
	TokenSignIn
	GoogleSignIn
)

type AccessLevel int32

const (
	Owner AccessLevel = iota
	Admin
	Moderator
	ReadAndWrite
	ReadOnly
	NoAccess
)

type Token string

type User struct {
	Id          uuid.UUID
	DisplayName string
	Created     time.Time
}

type UserConnection struct {
	UserId      uuid.UUID
	SignInType  SignInType
	AccountId   string
	AuthDetails *string
}

type UserStoreReader interface {
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserConnection(ctx context.Context, signInType SignInType, accountId string) (*UserConnection, error)
}

type UserStoreWriter interface {
	CreateUser(ctx context.Context, user User) error
	RemoveUser(ctx context.Context, id uuid.UUID) error
	SaveUserConnection(ctx context.Context, userConnection UserConnection) error
}

type UserStore interface {
	UserStoreReader
	UserStoreWriter
}

func (uc *UserConnection) IsValid() bool {
	if uc.UserId == uuid.Nil {
		return false
	}

	if uc.SignInType == None {
		return false
	}

	if uc.AccountId == "" {
		return false
	}

	return true
}
