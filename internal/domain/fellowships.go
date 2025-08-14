package domain

import (
	"context"

	"github.com/google/uuid"
)

type Fellowship struct {
	Id        uuid.UUID `json:"id"`
	CreatorId uuid.UUID `json:"creatorId"`
	Name      string    `json:"name"`
}

type FellowshipStoreReader interface {
	GetUserFellowships(ctx context.Context, userId uuid.UUID) ([]Fellowship, error)
	GetUserFellowshipIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
	GetUserAccessLevel(ctx context.Context, userId uuid.UUID, fellowshipId uuid.UUID) (AccessLevel, error)
}

type FellowshipStoreWriter interface {
	CreateFellowship(ctx context.Context, fellowship Fellowship) error
}

type FellowshipStore interface {
	FellowshipStoreReader
	FellowshipStoreWriter
}
