package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/feed"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/fellowships"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/middleware"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/users"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/db/postgresql"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("starting ToW Server...")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("getting config")
	config := getConfig()

	if err := appMain(ctx, logger, config); err != nil {
		logger.Error("start failed", "error", err)
		os.Exit(1)
	}

	logger.Info("finished")
}

func appMain(ctx context.Context, logger *slog.Logger, config *config) error {
	startTime := time.Now()
	logger.Info("initialising...")

	db, err := postgresql.NewDB(ctx, config, config)
	if err != nil {
		logger.Error("Failed to init DB", "error", err)
		os.Exit(1)
	}

	if db == nil {
		return fmt.Errorf("DB connection is invalid")
	}

	defer db.Close()

	logger.Info("preparing DB")
	err = postgresql.PrepareDB(ctx, db)
	if err != nil {
		logger.Error("Failed to prepare DB", "error", err)
		os.Exit(1)
	}

	logger.Info("setting up services")
	tokensService := service.NewTokensService(ctx, config, postgresql.NewKeyStore(config, db))
	mailService := service.NewMailService(config, config, logger)
	userService := service.NewUserService(postgresql.NewUserStore(db), tokensService, *mailService)
	fellowshipService := service.NewFellowshipService(postgresql.NewFellowshipStore(db))
	feedService := service.NewFeedService(postgresql.NewFeedStore(db), postgresql.NewFellowshipStore(db), postgresql.NewCircleStore(db))

	rt := api.ComposeRouters(users.NewRouter(userService), fellowships.NewRouter(fellowshipService), feed.NewRouter(feedService))

	middlewares := []api.MiddlewareFunc{middleware.AuthMiddleware(userService)}

	logger.Info("initialising server")
	server := api.NewServer(logger, config, healthCheck(db, startTime), middlewares, rt)
	return server.Start(ctx)
}

func healthCheck(db *sql.DB, startTime time.Time) api.HealthCheckerFunc {
	return func(ctx context.Context) ([]api.Health, error) {
		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("db ping: %w", err)
		}

		if row := db.QueryRowContext(ctx, "SELECT 1"); row.Err() != nil {
			return nil, fmt.Errorf("db read: %w", row.Err())
		}

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		systemDetails := map[string]any{
			"uptimeSeconds": time.Since(startTime).Seconds(),
			"goroutines":    runtime.NumGoroutine(),
			"memAllocMB":    float64(mem.Alloc) / 1024 / 1024,
			"memSysMB":      float64(mem.Sys) / 1024 / 1024,
			"gcCycles":      mem.NumGC,
		}

		return []api.Health{
			{Service: "ToW Server", Status: "OK", Time: time.Now().Local().String(), Details: systemDetails},
			{Service: "ToW DB", Status: "OK", Time: time.Now().Local().String(), Details: db.Stats()},
		}, nil
	}
}
