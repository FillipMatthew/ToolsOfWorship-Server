package service

import (
	"context"
	"fmt"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

func NewFeedService(store domain.FeedStore, fellowshipStore domain.FellowshipStore, circleStore domain.CircleStore) *FeedService {
	return &FeedService{feedStore: store, fellowshipStore: fellowshipStore, circleStore: circleStore}
}

type FeedService struct {
	feedStore       domain.FeedStore
	fellowshipStore domain.FellowshipStore
	circleStore     domain.CircleStore
}

func (f *FeedService) List(ctx context.Context, user domain.User, limit *int, before *time.Time, after *time.Time) ([]domain.Post, error) {
	fellowshipIDs, err := f.fellowshipStore.GetUserFellowshipIDs(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed get user fellowships: %v", err)
	}

	circleIDs, err := f.circleStore.GetUserCircleIDs(ctx, user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed get user circles: %v", err)
	}

	return f.feedStore.GetPosts(ctx, fellowshipIDs, circleIDs, limit, before, after)
}

func (f *FeedService) Post(ctx context.Context, user domain.User, post domain.Post) error {
	if post.FellowshipId != uuid.Nil && post.CircleId != uuid.Nil {
		return fmt.Errorf("cannot be post to both fellowship and circle")
	} else if post.FellowshipId == uuid.Nil && post.CircleId == uuid.Nil {
		return fmt.Errorf("post must be associated with either a fellowship or a circle")
	}

	if post.FellowshipId != uuid.Nil {
		if canPost, err := f.fellowshipStore.CanUserPostToFellowship(ctx, user.Id, post.FellowshipId); err != nil {
			return fmt.Errorf("unable to check user permissions for fellowship %s: %v", post.FellowshipId, err)
		} else if !canPost {
			return fmt.Errorf("user %s cannot post to fellowship %s", user.Id, post.FellowshipId)
		}
	}

	if post.CircleId != uuid.Nil {
		if canPost, err := f.circleStore.CanUserPostToCircle(ctx, user.Id, post.CircleId); err != nil {
			return fmt.Errorf("unable to check user permissions for circle %s: %v", post.CircleId, err)
		} else if !canPost {
			return fmt.Errorf("user %s cannot post to circle %s", user.Id, post.CircleId)
		}
	}

	return f.Post(ctx, user, post)
}
