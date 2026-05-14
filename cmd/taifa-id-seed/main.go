package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"taifa-id/internal/config"
	"taifa-id/internal/platform/password"
	"taifa-id/internal/platform/postgres"
	"taifa-id/internal/seed"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.Load()
	if cfg.Database.DSN == "" {
		logger.Error("TAIFA_ID_DATABASE_DSN is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dbPool, err := postgres.Open(ctx, postgres.Config{
		DSN:            cfg.Database.DSN,
		MinConns:       cfg.Database.MinConns,
		MaxConns:       cfg.Database.MaxConns,
		ConnectTimeout: cfg.Database.ConnectTimeout,
	})
	if err != nil {
		logger.Error("failed to open postgres", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	seedPassword := os.Getenv("TAIFA_ID_SEED_PASSWORD")
	if seedPassword == "" {
		seedPassword = seed.DefaultSeedPassword
	}

	runner := seed.NewRunner(
		dbPool,
		password.NewBcryptHasher(password.DefaultBcryptCost),
		seed.Config{
			DefaultPassword: seedPassword,
		},
	)

	result, err := runner.Run(ctx)
	if err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}

	logger.Info(
		"seed completed",
		"persons", result.Persons,
		"organizations", result.Organizations,
		"capabilities", result.Capabilities,
		"memberships", result.Memberships,
		"roles", result.Roles,
		"credentials", result.Credentials,
		"audit_events", result.AuditEvents,
	)

	fmt.Println("Seed credentials use usernames ending in .seed")
	fmt.Println("Default seed password:", seedPassword)
}
