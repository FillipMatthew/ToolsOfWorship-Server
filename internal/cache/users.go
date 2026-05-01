package cache

import (
	"context"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

type UserStore struct {
	inner     domain.UserStore
	userCache *Cache[uuid.UUID, *domain.User]
}

func NewUserStore(inner domain.UserStore, ttl time.Duration) *UserStore {
	return &UserStore{
		inner:     inner,
		userCache: New[uuid.UUID, *domain.User](ttl),
	}
}

func (s *UserStore) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if user, ok := s.userCache.Get(id); ok {
		return user, nil
	}

	user, err := s.inner.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	s.userCache.Set(id, user)
	return user, nil
}

func (s *UserStore) GetUserConnection(ctx context.Context, signInType domain.SignInType, accountId string) (*domain.UserConnection, error) {
	return s.inner.GetUserConnection(ctx, signInType, accountId)
}

func (s *UserStore) CreateUser(ctx context.Context, user domain.User) error {
	if err := s.inner.CreateUser(ctx, user); err != nil {
		return err
	}

	s.userCache.Delete(user.Id)
	return nil
}

func (s *UserStore) RemoveUser(ctx context.Context, id uuid.UUID) error {
	if err := s.inner.RemoveUser(ctx, id); err != nil {
		return err
	}

	s.userCache.Delete(id)
	return nil
}

func (s *UserStore) SaveUserConnection(ctx context.Context, userConnection domain.UserConnection) error {
	return s.inner.SaveUserConnection(ctx, userConnection)
}
