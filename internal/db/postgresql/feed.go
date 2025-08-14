package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func NewFeedStore(db *sql.DB) *FeedStore {
	return &FeedStore{db: db}
}

type FeedStore struct {
	db *sql.DB
}

func (f *FeedStore) GetPosts(ctx context.Context, fellowshipIDs []uuid.UUID, circleIDs []uuid.UUID, limit *int, before *time.Time, after *time.Time) ([]domain.Post, error) {
	var (
		args       []any
		conditions []string
	)

	if len(fellowshipIDs) == 0 && len(circleIDs) == 0 {
		return nil, fmt.Errorf("at least one fellowshipID or circleID must be provided")
	}

	query := "SELECT id, authorId, fellowshipId, circleId, posted, heading, article FROM Posts WHERE fellowshipId = ANY($1) OR circleId = ANY($2)"
	args = append(args, pq.Array(fellowshipIDs))
	args = append(args, pq.Array(circleIDs))

	if before != nil {
		conditions = append(conditions, fmt.Sprintf("posted < $%d", len(args)+1))
		args = append(args, *before)
	}

	if after != nil {
		conditions = append(conditions, fmt.Sprintf("posted > $%d", len(args)+1))
		args = append(args, *after)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	actualLimit := 10 // default limit
	if limit != nil {
		actualLimit = max(min(*limit, 1000), 1) // enforce a maximum limit and a minimum of 1
	}

	query += fmt.Sprintf(" ORDER BY posted DESC LIMIT %d", actualLimit)

	rows, err := f.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.Post, 0, actualLimit)

	for rows.Next() {
		post := domain.Post{}
		err := rows.Scan(&post.Id, &post.AuthorId, &post.FellowshipId, &post.CircleId, &post.Posted, &post.Heading, &post.Article)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (f *FeedStore) CreatePost(ctx context.Context, post domain.Post) error {
	if post.Id == uuid.Nil {
		panic("invalid post id")
	}

	_, err := f.db.ExecContext(ctx, "INSERT INTO Posts (id, authorId, fellowshipId, circleId, posted, heading, article) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		post.Id, post.AuthorId, post.FellowshipId, post.CircleId, post.Posted, post.Heading, post.Article)

	return err
}
