package model

import (
	"time"

	"github.com/google/uuid"
)

type PasswordReset struct {
	ID uuid.UUID `db:"id"`

	UserID uuid.UUID `db:"user_id"`

	Token string `db:"token"`

	ExpiresAt time.Time `db:"expires_at"`

	UsedAt *time.Time `db:"used_at"`

	CreatedAt time.Time `db:"created_at"`
}