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

type CircleMember struct {
	CircleId uuid.UUID
	UserId   uuid.UUID
	Access   AccessLevel
}

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
	GetCircleMembers(ctx context.Context, circleId uuid.UUID) ([]CircleMember, error)
}

type CircleStoreWriter interface {
	CreateCircle(ctx context.Context, circle Circle) error
	AddCircleMember(ctx context.Context, member CircleMember) error
}

type CircleStore interface {
	CircleStoreReader
	CircleStoreWriter
}
