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
	Id           uuid.UUID  `json:"id"`
	Creator      uuid.UUID  `json:"creatorId"`
	FellowshipId uuid.UUID  `json:"fellowshipId"`
	Name         string     `json:"name"`
	Type         CircleType `json:"type"`
}

type CircleStoreReader interface {
	GetUserCircles(ctx context.Context, userId uuid.UUID) ([]Circle, error)
	GetUserCircleIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
	GetUserAccessLevel(ctx context.Context, userId uuid.UUID, circleId uuid.UUID) (AccessLevel, error)
}

type CircleStoreWriter interface {
	CreateCircle(ctx context.Context, circle Circle) error
}

type CircleStore interface {
	CircleStoreReader
	CircleStoreWriter
}
