package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID uuid.UUID `db:"id"`

	UserID uuid.UUID `db:"user_id"`

	Token string `db:"token"`

	ExpiresAt time.Time `db:"expires_at"`

	RevokedAt *time.Time `db:"revoked_at"`

	IPAddress *string `db:"ip_address"`

	UserAgent *string `db:"user_agent"`

	CreatedAt time.Time `db:"created_at"`
}
