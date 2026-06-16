// Package model defines the database-level structs for auth-service.
// These map directly to PostgreSQL table rows.
// They are NOT the same as API request/response structs — keep them separate.
package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Username string    `db:"username"`

	Role   string `db:"role"`
	Status string `db:"status"`

	AvatarURL *string `db:"avatar_url"`

	IsVerified bool `db:"is_verified"`

	DeletedAt *time.Time `db:"deleted_at"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
