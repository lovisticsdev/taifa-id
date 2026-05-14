package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Environment string
	HTTP        HTTPConfig
	Database    DatabaseConfig
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN            string
	MinConns       int32
	MaxConns       int32
	ConnectTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName: getEnv("TAIFA_ID_SERVICE_NAME", "taifa-id"),
		Environment: getEnv("TAIFA_ID_ENV", "local"),
		HTTP: HTTPConfig{
			Addr:            getEnv("TAIFA_ID_HTTP_ADDR", ":8080"),
			ReadTimeout:     getDurationEnv("TAIFA_ID_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    getDurationEnv("TAIFA_ID_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getDurationEnv("TAIFA_ID_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDurationEnv("TAIFA_ID_HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			DSN:            getEnv("TAIFA_ID_DATABASE_DSN", ""),
			MinConns:       getInt32Env("TAIFA_ID_DATABASE_MIN_CONNS", 1),
			MaxConns:       getInt32Env("TAIFA_ID_DATABASE_MAX_CONNS", 5),
			ConnectTimeout: getDurationEnv("TAIFA_ID_DATABASE_CONNECT_TIMEOUT", 5*time.Second),
		},
	}

	if cfg.ServiceName == "" {
		return Config{}, fmt.Errorf("TAIFA_ID_SERVICE_NAME must not be empty")
	}

	if cfg.HTTP.Addr == "" {
		return Config{}, fmt.Errorf("TAIFA_ID_HTTP_ADDR must not be empty")
	}

	if cfg.Database.MaxConns < cfg.Database.MinConns {
		return Config{}, fmt.Errorf("TAIFA_ID_DATABASE_MAX_CONNS must be greater than or equal to TAIFA_ID_DATABASE_MIN_CONNS")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration
	}

	seconds, err := strconv.Atoi(value)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}

	return fallback
}

func getInt32Env(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}

	return int32(parsed)
}
