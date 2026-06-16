package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preloved-services/auth-service/internal/model"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(
	db *pgxpool.Pool,
) *RefreshTokenRepository {

	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	token *model.RefreshToken,
) error {

	query := `
	INSERT INTO auth.refresh_tokens (
		user_id,
		token,
		expires_at,
		ip_address,
		user_agent
	)
	VALUES (
		$1,$2,$3,$4,$5
	)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.IPAddress,
		token.UserAgent,
	)

	return err
}

func (r *RefreshTokenRepository) GetByHash(
	ctx context.Context,
	tokenHash string,
) (*model.RefreshToken, error) {

	query := `
	SELECT
		id,
		user_id,
		token,
		expires_at,
		revoked_at,
		ip_address,
		user_agent,
		created_at
	FROM auth.refresh_tokens
	WHERE token = $1
	LIMIT 1
	`

	var token model.RefreshToken

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.IPAddress,
		&token.UserAgent,
		&token.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenHash string,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE auth.refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		`,
		tokenHash,
	)

	return err
}

func (r *RefreshTokenRepository) DeleteExpired(
	ctx context.Context,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		DELETE FROM auth.refresh_tokens
		WHERE expires_at < NOW()
		`,
	)

	return err
}

func (r *RefreshTokenRepository) GetActiveSessions(
	ctx context.Context,
	userID string,
) ([]model.RefreshToken, error) {

	query := `
	SELECT
		id,
		user_id,
		token,
		expires_at,
		revoked_at,
		ip_address,
		user_agent,
		created_at
	FROM auth.refresh_tokens
	WHERE user_id = $1
	AND revoked_at IS NULL
	AND expires_at > NOW()
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.RefreshToken

	for rows.Next() {

		var session model.RefreshToken

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Token,
			&session.ExpiresAt,
			&session.RevokedAt,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

