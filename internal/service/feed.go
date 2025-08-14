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

	if len(fellowshipIDs) == 0 && len(circleIDs) == 0 {
		return []domain.Post{}, nil
	}

	return f.feedStore.GetPosts(ctx, fellowshipIDs, circleIDs, limit, before, after)
}

func (f *FeedService) Post(ctx context.Context, user domain.User, fellowshipId uuid.UUID, circleId uuid.UUID, heading string, article string) error {
	if fellowshipId != uuid.Nil && circleId != uuid.Nil {
		return fmt.Errorf("cannot be post to both fellowship and circle")
	} else if fellowshipId == uuid.Nil && circleId == uuid.Nil {
		return fmt.Errorf("post must be associated with either a fellowship or a circle")
	}

	if fellowshipId != uuid.Nil {
		if accessLevel, err := f.fellowshipStore.GetUserAccessLevel(ctx, user.Id, fellowshipId); err != nil {
			return fmt.Errorf("unable to check user permissions for fellowship %s: %v", fellowshipId, err)
		} else if !canPost(accessLevel) {
			return fmt.Errorf("user %s cannot post to fellowship %s", user.Id, fellowshipId)
		}
	}

	if circleId != uuid.Nil {
		if accessLevel, err := f.circleStore.GetUserAccessLevel(ctx, user.Id, circleId); err != nil {
			return fmt.Errorf("unable to check user permissions for circle %s: %v", circleId, err)
		} else if !canPost(accessLevel) {
			return fmt.Errorf("user %s cannot post to circle %s", user.Id, circleId)
		}
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate post ID: %v", err)
	}

	return f.feedStore.CreatePost(ctx, domain.Post{Id: uuid, AuthorId: user.Id, FellowshipId: fellowshipId, CircleId: circleId, Posted: time.Now(), Heading: heading, Article: article})
}

func canPost(accessLevel domain.AccessLevel) bool {
	if accessLevel == domain.Owner || accessLevel == domain.Admin || accessLevel == domain.Moderator || accessLevel == domain.ReadAndWrite {
		return true
	}

	return false
}
