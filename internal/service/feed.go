package service

import (
	"context"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

func NewFeedService(store domain.FeedStore) *FeedService {
	return &FeedService{feedStore: store}
}

type FeedService struct {
	feedStore domain.FeedStore
}

func (f *FeedService) List(ctx context.Context, user domain.User) ([]domain.Post, error) {
	return nil, nil
}

func (f *UserService) Post(ctx context.Context, user domain.User, post domain.Post) error {
	return nil
}
