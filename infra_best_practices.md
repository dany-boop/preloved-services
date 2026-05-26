# Preloved Services — Production Architecture & Developer Workflow Guide

This document establishes the architecture standards, naming conventions, and code guidelines for Preloved Services backend microservices.

---

## 1. Naming & Folder Conventions

### Naming Conventions
* **Docker Container Names**: `preloved_<service_name>` (e.g. `preloved_postgres`, `preloved_redis`)
* **Docker Network Name**: `preloved_net` (single bridge network)
* **Docker Volume Names**: `preloved_<database>_data` (e.g. `preloved_postgres_data`)
* **Environment Variables**: UPPERCASE with prefixes specifying the target system (e.g. `POSTGRES_USER`, `REDIS_PASSWORD`, `JWT_SECRET`).

### Recommended Project Folder Structure
For each Go microservice (e.g. `services/auth-service`), we enforce the standard Go project layout:
```text
services/auth-service/
├── cmd/
│   └── main.go                 # Entrypoint: parses config, initializes dependencies, starts listeners
├── config/
│   └── config.go               # Standardized config loader (struct mapping .env values)
├── internal/                   # Private application code (not importable by other services)
│   ├── handlers/               # REST API / HTTP controllers (Gin or Fiber)
│   ├── models/                 # DB models, schema structs
│   ├── repositories/           # Database abstraction layers (data access objects)
│   └── services/               # Core business logic
├── go.mod
└── go.sum
```

---

## 2. Migration Strategy & Dirty-State Recovery

We use **golang-migrate** for all database migrations.

### Best Practices
1. **Never modify existing migration files**: If a schema changes, write a *new* migration (`migrate create -ext sql ...`).
2. **Always write reversible migrations**: Every `0000XX_*.up.sql` must have a matching `0000XX_*.down.sql` that cleanly drops or alters the schema back.
3. **Avoid raw SQL command-line executions**: Never run manual `CREATE TABLE` commands. Track all schema changes in code.

### Reclaiming Dirty State
If a migration fails midway (e.g., syntax error), PostgreSQL will mark that migration as "dirty". Go applications will refuse to boot or migrate.
1. Fix the SQL error in the migration file.
2. Force the migration version back to the last known clean state:
   ```bash
   # If migration 4 failed, force version to 3:
   make migrate-force v=3
   ```
3. Re-run the migrations:
   ```bash
   make migrate-up
   ```

---

## 3. Reconnect-Safe Database Pool (Go & pgxpool)

To prevent the microservice from crashing if PostgreSQL restarts or drops connection momentarily, configure the connection pool with healthchecks and write a reconnect-safe initialization function using exponential backoff.

### Recommended Postgres Connection Blueprint
```go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ConnectPostgresPool initializes a reconnect-safe, pooled connection to PostgreSQL.
func ConnectPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres DSN: %w", err)
	}

	// Performance & Resilience Tuning
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	var pool *pgxpool.Pool
	maxRetries := 5
	backoff := 2 * time.Second

	for i := 1; i <= maxRetries; i++ {
		log.Info().Msgf("Attempting to connect to PostgreSQL (attempt %d/%d)...", i, maxRetries)
		
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			// Verify connection via Ping
			if err = pool.Ping(ctx); err == nil {
				log.Info().Msg("PostgreSQL connection pool established and verified.")
				return pool, nil
			}
		}

		log.Warn().Err(err).Msgf("PostgreSQL connection failed. Retrying in %v...", backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return nil, fmt.Errorf("could not connect to PostgreSQL after %d retries", maxRetries)
}
```

---

## 4. Graceful Shutdown Foundation

Microservices must terminate cleanly by listening for `SIGINT` (Ctrl+C) and `SIGTERM` (Docker/K8s stop signal), allowing active HTTP requests to complete, and closing database pools.

### Recommended Graceful Shutdown Blueprint
```go
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func run(pool *pgxpool.Pool, router http.Handler) {
	srv := &http.Server{
		Addr:    ":8001",
		Handler: router,
	}

	// Channel to listen for interrupt signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in background
	go func() {
		log.Info().Msgf("HTTP Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP Server failed")
		}
	}()

	// Block until a signal is received
	sig := <-shutdownChan
	log.Info().Msgf("Received signal %s. Beginning graceful shutdown...", sig)

	// Set a timeout context for active connections to drain
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Stop accepting new HTTP requests
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown before finishing active requests")
	} else {
		log.Info().Msg("HTTP Server closed cleanly.")
	}

	// 2. Close Database Pools
	log.Info().Msg("Closing PostgreSQL connection pool...")
	pool.Close()
	log.Info().Msg("PostgreSQL connection pool closed.")

	// 3. Close other systems (Redis client, RabbitMQ channels) if active...

	log.Info().Msg("Shutdown procedure complete.")
}
```

---

## 5. Structured Logging Foundation (Zerolog)

Microservices must log output in structured JSON formats to enable log collectors (like Elasticsearch or Grafana Loki) to index them. Local development should pretty-print logs for developer convenience.

### Logger Initialization Blueprint
```go
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger sets up the global logger based on environment.
func InitLogger(env string, logLevel string) {
	// Set log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Pretty printing for local dev, raw JSON for production/staging
	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	log.Info().Msgf("Logger initialized at %s level in %s mode", level, env)
}
```

### Correlation Middleware (Gin Example)
To trace requests across services, attach a Request ID and Correlation ID to every request context and inject them into every log:
```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		requestID := uuid.New().String()

		// Store in response headers
		c.Header("X-Correlation-ID", correlationID)
		c.Header("X-Request-ID", requestID)

		// Set variables in context for logging
		c.Set("CorrelationID", correlationID)
		c.Set("RequestID", requestID)

		// Create logger sub-context
		ctxLogger := log.With().
			Str("correlation_id", correlationID).
			Str("request_id", requestID).
			Logger()

		// Update context logger
		c.Request = c.Request.WithContext(ctxLogger.WithContext(c.Request.Context()))

		c.Next()
	}
}
```
