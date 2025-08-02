package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

func NewFeedStore(db *sql.DB) *FeedStore {
	return &FeedStore{db: db}
}

type FeedStore struct {
	db *sql.DB
}

func (f *FeedStore) GetPosts(ctx context.Context, id uuid.UUID, limit *int, before *time.Time, after *time.Time) ([]domain.Post, error) {
	return nil, nil
}

func (f *FeedStore) CreatePost(ctx context.Context, post *domain.Post) error {
	return nil
}
