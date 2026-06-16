package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preloved-services/auth-service/internal/model"
)

type EmailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(
	db *pgxpool.Pool,
) *EmailVerificationRepository {

	return &EmailVerificationRepository{
		db: db,
	}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, verification *model.EmailVerification) error {
	query := `
	INSERT INTO auth.email_verifications(
	user_id,
	token,
	expires_at
	)VALUES(
	$1, $2, $3
	)	
`

	_, err := r.db.Exec(ctx, query, verification.UserID, verification.Token, verification.ExpiresAt)
	return err
}
