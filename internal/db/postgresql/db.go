package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/config"
	"github.com/lib/pq"
)

func NewDB(ctx context.Context, dbConfig config.DatabaseConfig, poolConfig config.DatabasePoolConfig) (*sql.DB, error) {
	useSSL := "enable"
	if !dbConfig.UseSSL() {
		useSSL = "disable"
	}

	postgresConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.GetHost(), dbConfig.GetPort(), dbConfig.GetUser(), dbConfig.GetPassword(), "postgres", useSSL)

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
	row := db.QueryRowContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", dbConfig.GetName())

	var exists int
	err = row.Scan(&exists)
	if err == sql.ErrNoRows {
		queryStr := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbConfig.GetName()))
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
		dbConfig.GetHost(), dbConfig.GetPort(), dbConfig.GetUser(), dbConfig.GetPassword(), dbConfig.GetName(), useSSL)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot connect to database: %v", err)
	}

	if poolConfig != nil {
		if n := poolConfig.GetMaxOpenConns(); n > 0 {
			db.SetMaxOpenConns(n)
		}
		if n := poolConfig.GetMaxIdleConns(); n > 0 {
			db.SetMaxIdleConns(n)
		}
		if d := poolConfig.GetConnMaxLifetime(); d > 0 {
			db.SetConnMaxLifetime(d)
		}
	}

	return db, nil
}

func PrepareDB(ctx context.Context, db *sql.DB) error {
	return RunMigrations(ctx, db)
}
