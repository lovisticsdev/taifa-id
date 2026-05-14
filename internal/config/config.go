package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Environment string
	HTTP        HTTPConfig
	Database    DatabaseConfig
	Auth        AuthConfig
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

type AuthConfig struct {
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
	JWTTTL      time.Duration
}

func Load() Config {
	return Config{
		ServiceName: envString("TAIFA_ID_SERVICE_NAME", "taifa-id"),
		Environment: envString("TAIFA_ID_ENV", "local"),
		HTTP: HTTPConfig{
			Addr:            envString("TAIFA_ID_HTTP_ADDR", ":8080"),
			ReadTimeout:     envDuration("TAIFA_ID_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("TAIFA_ID_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("TAIFA_ID_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("TAIFA_ID_HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			DSN:            envString("TAIFA_ID_DATABASE_DSN", ""),
			MinConns:       envInt32("TAIFA_ID_DATABASE_MIN_CONNS", 1),
			MaxConns:       envInt32("TAIFA_ID_DATABASE_MAX_CONNS", 5),
			ConnectTimeout: envDuration("TAIFA_ID_DATABASE_CONNECT_TIMEOUT", 5*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:   envString("TAIFA_ID_JWT_SECRET", ""),
			JWTIssuer:   envString("TAIFA_ID_JWT_ISSUER", "taifa-id"),
			JWTAudience: envString("TAIFA_ID_JWT_AUDIENCE", "taifa-republic"),
			JWTTTL:      envDuration("TAIFA_ID_JWT_TTL", time.Hour),
		},
	}
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func envInt32(key string, fallback int32) int32 {
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
