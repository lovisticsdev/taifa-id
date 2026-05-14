package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"taifa-id/internal/config"
	"taifa-id/internal/platform/password"
	"taifa-id/internal/platform/postgres"
	"taifa-id/internal/seed"
)

const defaultSeedTimeout = 5 * time.Minute

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.Load()
	if cfg.Database.DSN == "" {
		logger.Error("TAIFA_ID_DATABASE_DSN is required")
		os.Exit(1)
	}

	seedTimeout := envDuration("TAIFA_ID_SEED_TIMEOUT", defaultSeedTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), seedTimeout)
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

	seedPassword, seedPasswordFromEnv := seedPassword()

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
		"timeout", seedTimeout.String(),
		"seed_password_from_env", seedPasswordFromEnv,
	)

	logger.Info("seed credential usernames use the .seed suffix")
	logger.Info("seed password value was not printed")
}

func seedPassword() (string, bool) {
	value := os.Getenv("TAIFA_ID_SEED_PASSWORD")
	if value == "" {
		return seed.DefaultSeedPassword, false
	}

	return value, true
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}