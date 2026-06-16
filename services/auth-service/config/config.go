// Package config loads environment variables into a typed Config struct.
// Every service has its own config — only loads what IT needs.
//
// USAGE in cmd/main.go:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal().Err(err).Msg("failed to load config")
//	}
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the auth-service.
// Values come from environment variables (loaded from .env in dev).
type Config struct {
	// App
	AppEnv      string // "development" | "production"
	AppLogLevel string // "debug" | "info" | "warn" | "error"

	// Server
	Port string // ":8001"

	// PostgreSQL — stores users, tokens, verifications
	PostgresDSN string // full connection string

	// Redis — stores sessions, rate limit counters, refresh token blocklist
	RedisAddr     string // "localhost:6379"
	RedisPassword string

	// JWT
	JWTSecret           string // must be 32+ chars in production
	JWTAccessExpiryMin  int    // minutes until access token expires (default: 15)
	JWTRefreshExpiryDay int    // days until refresh token expires (default: 7)
}

func Load() (*Config, error) {
	// Try loading .env — only needed locally.
	// godotenv.Load() silently does nothing if the file doesn't exist.
	_ = godotenv.Load("../../.env") // relative to services/auth-service/
	_ = godotenv.Load(".env")       // fallback: .env next to the binary

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppLogLevel: getEnv("APP_LOG_LEVEL", "debug"),
		Port:        ":" + getEnv("AUTH_SERVICE_PORT", "8001"),

		// Build the Postgres DSN from individual parts
		// DBeaver connects with these same values!
		PostgresDSN: buildPostgresDSN(),

		RedisAddr:     getEnv("REDIS_HOST", "localhost") + ":" + getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "pspassword"),

		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTAccessExpiryMin:  getEnvInt("JWT_ACCESS_EXPIRY_MIN", 15),
		JWTRefreshExpiryDay: getEnvInt("JWT_REFRESH_EXPIRY_DAY", 7),
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Build the Postgres DSN for the auth-service.
func buildPostgresDSN() string {

	user := getEnv("POSTGRES_USER", "aguser")
	pass := getEnv("POSTGRES_PASSWORD", "agpassword")
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	db := getEnv("POSTGRES_DB", "antigravity")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, db,
	)
}

// validate checks that required config values are set.
func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required — set it in .env (min 32 chars)")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	return nil
}

// Helper to read string env var or return default.
func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

// Helper to read int env var or return default.
func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}
