package domain

import (
	"context"

	"github.com/google/uuid"
)

type FellowshipMember struct {
	FellowshipId uuid.UUID
	UserId       uuid.UUID
	Access       AccessLevel
}

type Fellowship struct {
	Id        uuid.UUID `json:"id"`
	CreatorId uuid.UUID `json:"creatorId"`
	Name      string    `json:"name"`
}

type FellowshipStoreReader interface {
	GetUserFellowships(ctx context.Context, userId uuid.UUID) ([]Fellowship, error)
	GetUserFellowshipIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
	GetUserAccessLevel(ctx context.Context, userId uuid.UUID, fellowshipId uuid.UUID) (AccessLevel, error)
	GetFellowshipMembers(ctx context.Context, fellowshipId uuid.UUID) ([]FellowshipMember, error)
}

type FellowshipStoreWriter interface {
	CreateFellowship(ctx context.Context, fellowship Fellowship) error
	AddFellowshipMember(ctx context.Context, member FellowshipMember) error
}

type FellowshipStore interface {
	FellowshipStoreReader
	FellowshipStoreWriter
}
