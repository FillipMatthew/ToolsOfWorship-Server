package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/users"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/db/postgresql"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := log.New(os.Stdout, "ToW-Server: ", log.LstdFlags)
	config := getConfig()

	if err := appMain(ctx, logger, config); err != nil {
		logger.Fatalf("start failed: %v", err)
	}

	logger.Println("finished")
}

func appMain(ctx context.Context, logger *log.Logger, config *config) error {
	db, err := postgresql.NewDB(ctx, config)
	if err != nil {
		return err
	}

	defer db.Close()

	err = postgresql.PrepareDB(ctx, db)
	if err != nil {
		return err
	}

	tokensService := service.NewTokensService(ctx, config, postgresql.NewKeyStore(config, db))
	mailService := service.NewMailService(config, config)
	userService := service.NewUserService(postgresql.NewUserStore(db), *tokensService, *mailService)

	rt := api.ComposeRouters(users.NewRouter(userService))

	server := api.NewServer(logger, config, healthCheck(db), rt)
	return server.Start(ctx)
}

func healthCheck(db *sql.DB) api.HealthCheckerFunc {
	return func(ctx context.Context) ([]api.Health, error) {
		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("db ping: %w", err)
		}

		if row := db.QueryRowContext(ctx, "SELECT 1"); row.Err() != nil {
			return nil, fmt.Errorf("db read: %w", row.Err())
		}

		return []api.Health{
			{Service: "ToW Server", Status: "OK", Time: time.Now().Local().String()},
			{Service: "ToW DB", Status: "OK", Time: time.Now().Local().String(), Details: db.Stats()},
		}, nil
	}
}
