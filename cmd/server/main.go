package main

import (
	"context"
	"log/slog"
	"os"

	"net/http"
	"qback/internal/config"
	"qback/internal/database"
	"qback/internal/handlers"
	"qback/internal/middleware"
	"qback/internal/repository"
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
	slog.Info("configuration loaded", "port", cfg.Port)

	if err := database.RunMigrations(cfg); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected", "max_conns", 10, "min_conns", 2)

	userRepo := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	userHandler := handlers.NewUserHandler(userRepo)
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/signup", authHandler.Signup)
	mux.HandleFunc("/api/login", authHandler.Login)

	mux.Handle("/api/user/me", authMiddleware(http.HandlerFunc(userHandler.GetMe)))

	slog.Info("qback server starting...", "port", cfg.Port)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
