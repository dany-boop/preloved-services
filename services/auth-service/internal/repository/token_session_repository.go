package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	blacklistPrefix = "blacklist:"
	sessionPrefix   = "session:"
)

type TokenSessionRepository struct {
	redis *redis.Client
}

func NewTokenSessionRepository(redis *redis.Client) *TokenSessionRepository {
	return &TokenSessionRepository{redis: redis}
}

// BlacklistAccessToken stores a JWT until its expiration.
func (r *TokenSessionRepository) BlacklistToken(
	ctx context.Context,
	token string,
	expiresAt time.Time,
) error {

	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		return nil
	}

	key := blacklistPrefix + token

	return r.redis.Set(ctx, key, "1", ttl).Err()

}

// IsAccessTokenBlacklisted checks if JWT was revoked.
func (r *TokenSessionRepository) IsAccessTokenBlacklisted(
	ctx context.Context,
	token string,
) (bool, error) {
	key := blacklistPrefix + token
	_, err := r.redis.Get(ctx, key).Result()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *TokenSessionRepository) CreateUserSession(
	ctx context.Context,
	userID string,
	payload any,
	ttl time.Duration,
) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	key := sessionPrefix + userID
	return r.redis.Set(ctx, key, bytes, ttl).Err()
}

func (r *TokenSessionRepository) GetSession(
	ctx context.Context,
	userID string,
	dest any,
) error {
	key := sessionPrefix + userID
	value, err := r.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(value, dest)
}

func (r *TokenSessionRepository) DeleteUserSession(
	ctx context.Context,
	userID string,
) error {
	pattern := fmt.Sprintf("%s%s", sessionPrefix, userID)

	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return ErrNotFound
	}
	return r.redis.Del(ctx, keys...).Err()
}
