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
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %v", err)
	}

	// Create DB if not existing
	queryStr := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", config.GetName())
	row := db.QueryRowContext(ctx, queryStr)
	if row.Err() != nil {
		queryStr = fmt.Sprintf("CREATE DATABASE %s", config.GetName())
		_, err := db.ExecContext(ctx, queryStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %v", row.Err())
		}
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
