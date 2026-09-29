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
	"qback/internal/models"
	"qback/internal/repository"
	"qback/internal/service"
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

	serviceRepo := repository.NewServiceRepository(db)
	serviceLogic := service.NewService(serviceRepo)
	serviceHandler := handlers.NewServiceHandler(serviceLogic)

	venueRepo := repository.NewVenueRepository(db)
	venueLogic := service.NewVenueService(venueRepo)
	venueHandler := handlers.NewVenueHandler(venueLogic)

	queueRepo := repository.NewQueueTicketRepository(db)
	queueLogic := service.NewQueueService(queueRepo, serviceRepo)
	tokenHandler := handlers.NewTokenHandler(queueLogic)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/signup", authHandler.Signup)
	mux.HandleFunc("/api/login", authHandler.Login)

	requireUser := middleware.RequireRole(string(models.RoleUser), string(models.RoleStaff), string(models.RoleAdmin))
	
	protectedUser := func(h http.Handler) http.Handler {
		return authMiddleware(requireUser(h))
	}

	mux.Handle("/api/user/me", protectedUser(http.HandlerFunc(userHandler.GetMe)))

	mux.HandleFunc("GET /api/services", serviceHandler.GetAll)
	mux.HandleFunc("GET /api/services/{id}", serviceHandler.GetByID)

	mux.HandleFunc("GET /api/venues", venueHandler.GetAll)
	mux.HandleFunc("GET /api/venues/{id}", venueHandler.GetByID)
	mux.HandleFunc("GET /api/venues/{id}/services", serviceHandler.GetByVenueID)

	mux.Handle("POST /api/tokens", protectedUser(http.HandlerFunc(tokenHandler.JoinQueue)))
	mux.Handle("GET /api/tokens/active", protectedUser(http.HandlerFunc(tokenHandler.GetActive)))
	mux.Handle("GET /api/tokens/history", protectedUser(http.HandlerFunc(tokenHandler.GetHistory)))
	mux.Handle("GET /api/tokens/{id}", protectedUser(http.HandlerFunc(tokenHandler.GetByID)))
	mux.Handle("POST /api/tokens/{id}/cancel", protectedUser(http.HandlerFunc(tokenHandler.Cancel)))

	slog.Info("server starting...", "port", cfg.Port)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
