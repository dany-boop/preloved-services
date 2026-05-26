// Package database provides connection helpers for all databases.
// Each service imports only the databases it needs.
//
// USAGE:
//   // In your service cmd/main.go:
//   db, err := database.NewPostgres(cfg.PostgresDSN)
//   mongo, err := database.NewMongo(cfg.MongoURI)
//   rdb, err := database.NewRedis(cfg.RedisAddr, cfg.RedisPassword)

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ──────────────────────────────────────────────
// PostgreSQL
// ──────────────────────────────────────────────

// NewPostgres creates a PostgreSQL connection pool.
// pgxpool is connection-pool aware — safe for concurrent use across goroutines.
//
// dsn format: "postgres://user:pass@host:5432/dbname"
func NewPostgres(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	// Tuning: adjust based on your server resources
	config.MaxConns = 25                      // max simultaneous connections
	config.MinConns = 5                       // keep 5 warm connections ready
	config.MaxConnLifetime = 30 * time.Minute // recycle old connections
	config.MaxConnIdleTime = 5 * time.Minute  // close idle connections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	// Verify the connection works
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

// ──────────────────────────────────────────────
// MongoDB
// ──────────────────────────────────────────────

// NewMongo creates a MongoDB client.
// uri format: "mongodb://user:pass@host:27017"
func NewMongo(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(20).           // max connections in pool
		SetMinPoolSize(5).            // keep warm connections
		SetConnectTimeout(5 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	// Verify the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return client, nil
}

// ──────────────────────────────────────────────
// Redis
// ──────────────────────────────────────────────

// NewRedis creates a Redis client.
// addr format: "localhost:6379"
func NewRedis(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,          // default DB
		PoolSize:     20,         // max connections
		MinIdleConns: 5,          // keep warm
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}
