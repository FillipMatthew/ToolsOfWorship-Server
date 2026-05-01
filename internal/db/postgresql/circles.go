package postgresql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
)

func NewCircleStore(db *sql.DB) *CircleStore {
	return &CircleStore{db: db}
}

type CircleStore struct {
	db *sql.DB
}

func (c *CircleStore) GetUserCircles(ctx context.Context, userId uuid.UUID) ([]domain.Circle, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT id, fellowshipId, name, type, creator FROM FellowshipCircles WHERE id in (SELECT circleId FROM CircleMembers WHERE userId=$1)", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	circles := make([]domain.Circle, 0)

	for rows.Next() {
		circle := domain.Circle{}
		err := rows.Scan(&circle.Id, &circle.FellowshipId, &circle.Name, &circle.Type, &circle.Creator)
		if err != nil {
			return nil, err
		}

		circles = append(circles, circle)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return circles, nil
}

func (c *CircleStore) GetUserCircleIDs(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT circleId FROM CircleMembers WHERE userId=$1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	circleIds := make([]uuid.UUID, 0)

	for rows.Next() {
		circleId := uuid.UUID{}
		err := rows.Scan(&circleId)
		if err != nil {
			return nil, err
		}

		circleIds = append(circleIds, circleId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return circleIds, nil
}

func (c *CircleStore) GetUserAccessLevel(ctx context.Context, userId uuid.UUID, circleId uuid.UUID) (domain.AccessLevel, error) {
	accessLevel := domain.NoAccess

	err := c.db.QueryRowContext(ctx, "SELECT access FROM CircleMembers WHERE userId=$1 AND circleId=$2", userId, circleId).Scan(&accessLevel)
	if err != nil {
		return domain.NoAccess, err
	}

	return accessLevel, nil
}

func (c *CircleStore) GetCircleMembers(ctx context.Context, circleId uuid.UUID) ([]domain.CircleMember, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT userId, access FROM CircleMembers WHERE circleId=$1", circleId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.CircleMember, 0)

	for rows.Next() {
		member := domain.CircleMember{CircleId: circleId}
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

func (c *CircleStore) CreateCircle(ctx context.Context, circle domain.Circle) error {
	return errors.New("not implemented")
}

func (c *CircleStore) AddCircleMember(ctx context.Context, member domain.CircleMember) error {
	_, err := c.db.ExecContext(ctx, "INSERT INTO CircleMembers (circleId, userId, access) VALUES ($1, $2, $3)", member.CircleId, member.UserId, member.Access)
	return err
}
