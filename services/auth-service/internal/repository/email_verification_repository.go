package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

func (r *EmailVerificationRepository) GetByToken(
	ctx context.Context,
	token string,
) (*model.EmailVerification, error) {

	query := `
	SELECT
		id,
		user_id,
		token,
		expires_at,
		used_at,
		created_at
	FROM auth.email_verifications
	WHERE token = $1
	LIMIT 1
	`

	var verification model.EmailVerification

	err := r.db.QueryRow(
		ctx,
		query,
		token,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.Token,
		&verification.ExpiresAt,
		&verification.UsedAt,
		&verification.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) MarkAsUsed(
	ctx context.Context,
	token string,
) error {

	query := `
	UPDATE auth.email_verifications
	SET used_at = NOW()
	WHERE token = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		token,
	)

	return err
}

func (r *EmailVerificationRepository) IsUsed(
	ctx context.Context,
	token string,
) (bool, error) {

	var used bool

	query := `
	SELECT EXISTS (
		SELECT 1
		FROM auth.email_verifications
		WHERE token = $1
		AND used_at IS NOT NULL
	)
	`

	err := r.db.QueryRow(
		ctx,
		query,
		token,
	).Scan(&used)

	return used, err
}

func (r *EmailVerificationRepository) DeleteByUserID(
	ctx context.Context,
	userID string,
) error {

	query := `
	DELETE FROM auth.email_verifications
	WHERE user_id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		userID,
	)

	return err
}
