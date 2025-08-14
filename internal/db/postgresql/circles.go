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
	return domain.NoAccess, errors.New("not implemented")
}

func (c *CircleStore) CreateCircle(ctx context.Context, circle domain.Circle) error {

	return errors.New("not implemented")
}
