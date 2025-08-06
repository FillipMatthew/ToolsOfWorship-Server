package domain

import (
	"context"

	"github.com/google/uuid"
)

type Fellowship struct {
	Id      uuid.UUID
	Creator uuid.UUID
	Name    string
}

type FellowshipStoreReader interface {
	GetUserFellowships(ctx context.Context, userId uuid.UUID) ([]Fellowship, error)
	GetUserFellowshipIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
	CanUserPostToFellowship(ctx context.Context, userId uuid.UUID, fellowshipId uuid.UUID) (bool, error)
}

type FellowshipStoreWriter interface {
	CreateFellowship(ctx context.Context, fellowship Fellowship) error
}

type FellowshipStore interface {
	FellowshipStoreReader
	FellowshipStoreWriter
}
