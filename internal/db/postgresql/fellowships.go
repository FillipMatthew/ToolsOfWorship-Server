package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

func NewFellowshipStore(db *sql.DB) *FellowshipStore {
	return &FellowshipStore{db: db}
}

type FellowshipStore struct {
	db *sql.DB
}

func (f *FellowshipStore) GetUserFellowships(ctx context.Context, userId uuid.UUID) ([]domain.Fellowship, error) {
	rows, err := f.db.QueryContext(ctx, "SELECT id, name, creator FROM Fellowships WHERE id in (SELECT fellowshipId FROM FellowshipMembers WHERE userId=$1)", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fellowships := make([]domain.Fellowship, 0)

	for rows.Next() {
		fellowship := domain.Fellowship{}
		err := rows.Scan(&fellowship.Id, &fellowship.Name, &fellowship.Creator)
		if err != nil {
			return nil, err
		}

		fellowships = append(fellowships, fellowship)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fellowships, nil
}

func (f *FellowshipStore) GetUserFellowshipIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	rows, err := f.db.QueryContext(ctx, "SELECT fellowshipId FROM FellowshipMembers WHERE userId=$1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fellowshipIds := make([]uuid.UUID, 0)

	for rows.Next() {
		fellowshipId := uuid.UUID{}
		err := rows.Scan(&fellowshipId)
		if err != nil {
			return nil, err
		}

		fellowshipIds = append(fellowshipIds, fellowshipId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fellowshipIds, nil
}

func (f *FellowshipStore) CanUserPostToFellowship(ctx context.Context, userID uuid.UUID, fellowshipID uuid.UUID) (bool, error) {
	return false, errors.New("not implemented")
}

func (f *FellowshipStore) CreateFellowship(ctx context.Context, fellowship domain.Fellowship) error {
	return errors.New("not implemented")
}
