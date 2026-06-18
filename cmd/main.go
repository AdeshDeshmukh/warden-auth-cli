package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/cli"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/config"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/logger"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/repository/sqlite"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/service"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0755); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	appLogger, err := logger.New(cfg.LogPath)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	db, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := sqlite.NewUserRepo(db)
	sessionRepo := sqlite.NewSessionRepo(db)
	auditRepo := sqlite.NewAuditRepo(db)

	authService := service.NewAuthService(userRepo, auditRepo, cfg, appLogger)
	sessionService := service.NewSessionService(sessionRepo, auditRepo, cfg, appLogger)
	totpService := service.NewTOTPService(userRepo, auditRepo, appLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sessionService.CleanExpired(ctx); err != nil {
		appLogger.Warn("failed to clean expired sessions on startup", "error", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
		db.Close()
		os.Exit(0)
	}()

	app, err := cli.NewApp(authService, sessionService, totpService)
	if err != nil {
		log.Fatalf("failed to initialize CLI: %v", err)
	}

	app.Run(ctx)
}