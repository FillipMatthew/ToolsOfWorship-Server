package service

import (
	"context"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
)

func NewFellowshipService(store domain.FellowshipStore) *FellowshipService {
	return &FellowshipService{fellowshipStore: store}
}

type FellowshipService struct {
	fellowshipStore domain.FellowshipStore
}

func (f *FellowshipService) List(ctx context.Context, user domain.User) ([]domain.Fellowship, error) {
	return f.fellowshipStore.GetUserFellowships(ctx, user.Id)
}
