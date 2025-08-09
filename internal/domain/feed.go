package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Id           uuid.UUID `json:"id"`
	AuthorId     uuid.UUID `json:"authorId"`
	FellowshipId uuid.UUID `json:"fellowshipId"`
	CircleId     uuid.UUID `json:"circleId"`
	Posted       time.Time `json:"posted"`
	Heading      string    `json:"heading"`
	Article      string    `json:"article"`
}

type FeedStoreReader interface {
	GetPosts(ctx context.Context, fellowshipIDs []uuid.UUID, circleIDs []uuid.UUID, limit *int, before *time.Time, after *time.Time) ([]Post, error)
}

type FeedStoreWriter interface {
	CreatePost(ctx context.Context, post Post) error
}

type FeedStore interface {
	FeedStoreReader
	FeedStoreWriter
}
