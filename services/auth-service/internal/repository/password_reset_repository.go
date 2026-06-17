package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preloved-services/auth-service/internal/model"
)

type PasswordResetRepository struct {
	db *pgxpool.Pool
}

func NewPasswordResetRepository(db *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, reset *model.PasswordReset) error {

	query := `INSERT INTO auth.password_resets (id,user_id,token,expires_at)VALUES ($1,$2,$3)`

	_, err := r.db.Exec(ctx, query, reset.Token, reset.UserID, reset.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) FindByToken(ctx context.Context, token string) (*model.PasswordReset, error) {
	query := `SELECT id,user_id,token,expires_at ,used_at,created_at FROM auth.password_resets WHERE token=$1 AND LIMIT 1`

	var reset model.PasswordReset
	err := r.db.QueryRow(ctx, query, token).Scan(&reset.Token, &reset.UserID, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find password reset: %w", err)
	}
	return &reset, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, token string) error {

	query := `UPDATE auth.password_resets SET used_at = NOW() WHERE token = $1`
	_, err := r.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {

	query := `DELETE FROM auth.password_resets WHERE expires_at < NOW()`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("delete expired password resets: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) RevokeAllUserResets(
	ctx context.Context,
	userID string,
) error
