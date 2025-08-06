package domain

import (
	"context"

	"github.com/google/uuid"
)

type CircleType int32

const (
	Notices CircleType = iota
	Prayer
)

type Circle struct {
	Id           uuid.UUID
	Creator      uuid.UUID
	FellowshipId uuid.UUID
	Name         string
	Type         CircleType
}

type CircleStoreReader interface {
	GetUserCircles(ctx context.Context, userId uuid.UUID) ([]Circle, error)
	GetUserCircleIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
	CanUserPostToCircle(ctx context.Context, userId uuid.UUID, circleId uuid.UUID) (bool, error)
}

type CircleStoreWriter interface {
	CreateCircle(ctx context.Context, circle Circle) error
}

type CircleStore interface {
	CircleStoreReader
	CircleStoreWriter
}
