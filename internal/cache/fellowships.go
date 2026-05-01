package cache

import (
	"context"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

type FellowshipStore struct {
	inner              domain.FellowshipStore
	fellowshipsCache   *Cache[uuid.UUID, []domain.Fellowship]
	fellowshipIDsCache *Cache[uuid.UUID, []uuid.UUID]
}

func NewFellowshipStore(inner domain.FellowshipStore, ttl time.Duration) *FellowshipStore {
	return &FellowshipStore{
		inner:              inner,
		fellowshipsCache:   New[uuid.UUID, []domain.Fellowship](ttl),
		fellowshipIDsCache: New[uuid.UUID, []uuid.UUID](ttl),
	}
}

func (s *FellowshipStore) GetUserFellowships(ctx context.Context, userId uuid.UUID) ([]domain.Fellowship, error) {
	if fellowships, ok := s.fellowshipsCache.Get(userId); ok {
		return fellowships, nil
	}

	fellowships, err := s.inner.GetUserFellowships(ctx, userId)
	if err != nil {
		return nil, err
	}

	s.fellowshipsCache.Set(userId, fellowships)
	return fellowships, nil
}

func (s *FellowshipStore) GetUserFellowshipIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	if ids, ok := s.fellowshipIDsCache.Get(userId); ok {
		return ids, nil
	}

	ids, err := s.inner.GetUserFellowshipIDs(ctx, userId)
	if err != nil {
		return nil, err
	}

	s.fellowshipIDsCache.Set(userId, ids)
	return ids, nil
}

func (s *FellowshipStore) GetUserAccessLevel(ctx context.Context, userId uuid.UUID, fellowshipId uuid.UUID) (domain.AccessLevel, error) {
	return s.inner.GetUserAccessLevel(ctx, userId, fellowshipId)
}

func (s *FellowshipStore) GetFellowshipMembers(ctx context.Context, fellowshipId uuid.UUID) ([]domain.FellowshipMember, error) {
	return s.inner.GetFellowshipMembers(ctx, fellowshipId)
}

func (s *FellowshipStore) CreateFellowship(ctx context.Context, fellowship domain.Fellowship) error {
	return s.inner.CreateFellowship(ctx, fellowship)
}

func (s *FellowshipStore) AddFellowshipMember(ctx context.Context, member domain.FellowshipMember) error {
	return s.inner.AddFellowshipMember(ctx, member)
}
