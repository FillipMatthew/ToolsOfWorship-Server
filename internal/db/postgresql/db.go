package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func NewDB(ctx context.Context, config config.DatabaseConfig) (*sql.DB, error) {
	useSSL := "enable"
	if !config.UseSSL() {
		useSSL = "disable"
	}

	postgresConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.GetHost(), config.GetPort(), config.GetUser(), config.GetPassword(), "postgres", useSSL)

	db, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		fmt.Printf("error opening database: %v", err)
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		fmt.Printf("cannot connect to database: %v", err)
		return nil, fmt.Errorf("cannot connect to database: %v", err)
	}

	// Create DB if not existing
	row := db.QueryRowContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", config.GetName())

	var exists int
	err = row.Scan(&exists)
	if err == sql.ErrNoRows {
		queryStr := fmt.Sprintf("CREATE DATABASE %s", config.GetName())
		_, err := db.ExecContext(ctx, queryStr)
		if err != nil {
			fmt.Printf("failed to create database: %v", err)
			return nil, fmt.Errorf("failed to create database: %v", err)
		}
	} else if err != nil {
		fmt.Printf("failed to check database state: %v", row.Err())
		return nil, fmt.Errorf("failed to check database state: %v", row.Err())
	}
	db.Close()

	// Open actual DB
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.GetHost(), config.GetPort(), config.GetUser(), config.GetPassword(), config.GetName(), useSSL)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot connect to database: %v", err)
	}

	return db, nil
}

func PrepareDB(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS SignKeys (id UUID PRIMARY KEY, key BYTEA NOT NULL, expiry TIMESTAMPTZ NOT NULL)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS EncKeys (id UUID PRIMARY KEY, key BYTEA NOT NULL, expiry TIMESTAMPTZ NOT NULL)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS Users (id UUID PRIMARY KEY, displayName VARCHAR(50) NOT NULL, isDeleted BOOLEAN DEFAULT FALSE NOT NULL)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS UserConnections (userId UUID NOT NULL, signInType INTEGER NOT NULL, accountId TEXT NOT NULL, authDetails TEXT)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS Posts (id UUID PRIMARY KEY, authorId UUID NOT NULL, fellowshipId UUID, circleId UUID, dateTime TIMESTAMPTZ NOT NULL, heading VARCHAR(80), article TEXT)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX idx_posts_datetime ON Posts(dateTime)")
	if err != nil {
		return err
	}

	return nil
}
