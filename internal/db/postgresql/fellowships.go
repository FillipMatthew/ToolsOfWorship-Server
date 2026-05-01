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
		err := rows.Scan(&fellowship.Id, &fellowship.Name, &fellowship.CreatorId)
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

func (f *FellowshipStore) GetUserAccessLevel(ctx context.Context, userId uuid.UUID, fellowshipId uuid.UUID) (domain.AccessLevel, error) {
	accessLevel := domain.NoAccess

	err := f.db.QueryRowContext(ctx, "SELECT access FROM FellowshipMembers WHERE userId=$1 AND fellowshipId=$2", userId, fellowshipId).Scan(&accessLevel)
	if err != nil {
		return domain.NoAccess, err
	}

	return accessLevel, nil
}

func (f *FellowshipStore) GetFellowshipMembers(ctx context.Context, fellowshipId uuid.UUID) ([]domain.FellowshipMember, error) {
	rows, err := f.db.QueryContext(ctx, "SELECT userId, access FROM FellowshipMembers WHERE fellowshipId=$1", fellowshipId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.FellowshipMember, 0)

	for rows.Next() {
		member := domain.FellowshipMember{FellowshipId: fellowshipId}
		if err := rows.Scan(&member.UserId, &member.Access); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (f *FellowshipStore) CreateFellowship(ctx context.Context, fellowship domain.Fellowship) error {
	return errors.New("not implemented")
}

func (f *FellowshipStore) AddFellowshipMember(ctx context.Context, member domain.FellowshipMember) error {
	_, err := f.db.ExecContext(ctx, "INSERT INTO FellowshipMembers (fellowshipId, userId, access) VALUES ($1, $2, $3)", member.FellowshipId, member.UserId, member.Access)
	return err
}
