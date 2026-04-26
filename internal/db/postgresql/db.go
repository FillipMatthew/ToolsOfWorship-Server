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
	_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS SignKeys (id UUID PRIMARY KEY, key BYTEA NOT NULL, expiry TIMESTAMPTZ NOT NULL)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS EncKeys (id UUID PRIMARY KEY, key BYTEA NOT NULL, expiry TIMESTAMPTZ NOT NULL)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS Users (id UUID PRIMARY KEY, displayName VARCHAR(50) NOT NULL, created TIMESTAMPTZ NOT NULL, isDeleted BOOLEAN DEFAULT FALSE NOT NULL)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS UserConnections (userId UUID NOT NULL REFERENCES Users(id), signInType INTEGER NOT NULL, accountId TEXT NOT NULL, authDetails TEXT, PRIMARY KEY (userId, signInType), UNIQUE (signInType, accountId))")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS Fellowships (id UUID PRIMARY KEY, name TEXT NOT NULL, creator UUID NOT NULL REFERENCES Users(id))")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS FellowshipMembers (fellowshipId UUID REFERENCES Fellowships(id), userId UUID REFERENCES Users(id), access INTEGER NOT NULL, PRIMARY KEY(fellowshipId, userId))")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS FellowshipCircles (id UUID PRIMARY KEY, fellowshipId UUID NOT NULL REFERENCES Fellowships(id), name TEXT NOT NULL, type INTEGER NOT NULL, creator UUID NOT NULL REFERENCES Users(id))")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS CircleMembers (circleId UUID REFERENCES FellowshipCircles(id), userId UUID REFERENCES Users(id), access INTEGER NOT NULL, PRIMARY KEY(circleId, userId))")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS Posts (id UUID PRIMARY KEY, authorId UUID NOT NULL REFERENCES Users(id), fellowshipId UUID, circleId UUID, posted TIMESTAMPTZ NOT NULL, heading VARCHAR(80), article TEXT)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_posts_posted ON Posts(posted)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_userconnections_accountid ON UserConnections(accountId)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_userconnections_userid ON UserConnections(userId)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_fellowshipmembers_userid ON FellowshipMembers(userId)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_circlemembers_userid ON CircleMembers(userId)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_posts_fellowshipid ON Posts(fellowshipId)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_posts_circleid ON Posts(circleId)")
	if err != nil {
		return err
	}

	return nil
}
