package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/sharooq/qback/internal/config"
	"github.com/sharooq/qback/internal/database"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded", "env", cfg.Env, "port", cfg.Port)

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected", "max_conns", 10, "min_conns", 2)

	slog.Info("qback server starting...", "port", cfg.Port)
}
