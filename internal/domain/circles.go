package domain

import (
	"context"

	"github.com/google/uuid"
)

type Circle struct {
	Id           uuid.UUID
	Creator      uuid.UUID
	FellowshipID uuid.UUID
	Name         string
}

type CircleStoreReader interface {
	GetUserCircles(ctx context.Context, userID uuid.UUID) ([]Circle, error)
	GetUserCircleIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	CanUserPostToCircle(ctx context.Context, userID uuid.UUID, circleID uuid.UUID) (bool, error)
}

type CircleStoreWriter interface {
	CreateCircle(ctx context.Context, circle Circle) error
}

type CircleStore interface {
	CircleStoreReader
	CircleStoreWriter
}
