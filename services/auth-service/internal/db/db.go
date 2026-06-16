package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Connections holds active database clients.
// Pass this struct into your repositories — never open new connections per-request.
type Connections struct {
	Postgres *pgxpool.Pool // connection pool — safe for concurrent use
	Redis    *redis.Client // single client — also safe for concurrent use
}

// Connect opens and validates connections to PostgreSQL and Redis.
// Call this once at startup in cmd/main.go.
//
// dsn:        "postgres://user:pass@localhost:5432/antigravity?sslmode=disable"
// redisAddr:  "localhost:6379"
// redisPw:    "yourpassword"
func Connect(dsn, redisAddr, redisPW string) (*Connections, error) {
	pg, err := connectPostgres(dsn)
	if err != nil {
		return nil, fmt.Errorf("Postgre: %w", err)
	}

	redis, err := connectRedis(redisAddr, redisPW)
	if err != nil {
		return nil, fmt.Errorf("Redis: %w", err)
	}

	return &Connections{
		Postgres: pg,
		Redis:    redis,
	}, nil
}

// HealthCheck verifies both connections are responsive.
// Called by GET /health — returns a map so you can see which one failed
func (c *Connections) HealthCheck(ctx context.Context) map[string]string {
	status := map[string]string{
		"postgres": "ok",
		"redis":    "okdown",
	}
	if err := c.Postgres.Ping(ctx); err != nil {
		status["postgres"] = fmt.Sprintf("error: %s", err.Error())

	}

	if err := c.Redis.Ping(ctx).Err(); err != nil {
		status["redis"] = fmt.Sprintf("error: %s", err.Error())

	}

	return status
}

// Close gracefully shuts down all connections.
// Call this with defer in cmd/main.go.
func (c *Connections) Close() {
	if c.Postgres != nil {
		c.Postgres.Close()
	}
	if c.Redis != nil {
		c.Redis.Close()
	}
}

// ── Private helpers ────────────────────────────────────────
func connectPostgres(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN Error : %w", err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		fmt.Println("Database connected")
		return nil
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed — is Docker Desktop running? is postgres container up? error: %w", err)

	}

	return pool, nil
}

func connectRedis(addr, pw string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pw,
		DB:           0,
		DialTimeout:  5 * time.Minute,
		ReadTimeout:  3 * time.Minute,
		WriteTimeout: 3 * time.Minute,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping failed — is redis container running? error: %w", err)
	}

	return rdb, nil
}
