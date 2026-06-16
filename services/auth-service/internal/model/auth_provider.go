package model

import (
	"time"

	"github.com/google/uuid"
)

type AuthProvider struct {
	ID uuid.UUID `db:"id"`

	UserID uuid.UUID `db:"user_id"`

	Provider string `db:"provider"`

	ProviderUserID *string `db:"provider_user_id"`

	PasswordHash *string `db:"password_hash"`

	WalletAddress *string `db:"wallet_address"`

	AccessToken *string `db:"access_token"`

	RefreshToken *string `db:"refresh_token"`

	CreatedAt time.Time `db:"created_at"`
}
