package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Id           uuid.UUID
	AuthorId     uuid.UUID
	FellowshipId *uuid.UUID
	CircleId     *uuid.UUID
	DateTime     time.Time
	Heading      string
	Article      string
}

type FeedStoreReader interface {
	GetPosts(ctx context.Context, id uuid.UUID, limit *int, before *time.Time, after *time.Time) ([]Post, error)
}

type FeedStoreWriter interface {
	CreatePost(ctx context.Context, post *Post) error
}

type FeedStore interface {
	FeedStoreReader
	FeedStoreWriter
}
