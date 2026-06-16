package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preloved-services/auth-service/internal/model"
)

type AuthProviderRepository struct {
	db *pgxpool.Pool
}

func NewAuthProviderRepository(
	db *pgxpool.Pool,
) *AuthProviderRepository {

	return &AuthProviderRepository{
		db: db,
	}
}

func (r *AuthProviderRepository) CreateLocalProvider(
	ctx context.Context,
	provider *model.AuthProvider,
) error {

	query := `
	INSERT INTO auth.auth_providers (
		user_id,
		provider,
		password_hash
	)
	VALUES (
		$1,
		'local',
		$2
	)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		provider.UserID,
		provider.PasswordHash,
	)

	return err
}

func (r *AuthProviderRepository) GetLocalProviderByEmail(
	ctx context.Context,
	email string,
) (*model.AuthProvider, error) {

	query := `
	SELECT
		ap.id,
		ap.user_id,
		ap.provider,
		ap.password_hash
	FROM auth.auth_providers ap
	INNER JOIN auth.users u
		ON u.id = ap.user_id
	WHERE
		u.email = $1
		AND ap.provider = 'local'
	LIMIT 1
	`

	var provider model.AuthProvider

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&provider.ID,
		&provider.UserID,
		&provider.Provider,
		&provider.PasswordHash,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (r *AuthProviderRepository) CreateGoogleProvider(
	ctx context.Context,
	provider *model.AuthProvider,
) error {

	query := `
	INSERT INTO auth.auth_providers (
		user_id,
		provider,
		provider_user_id
	)
	VALUES (
		$1,
		'google',
		$2
	)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		provider.UserID,
		provider.ProviderUserID,
	)

	return err
}
func (r *AuthProviderRepository) CreateWalletProvider(
	ctx context.Context,
	provider *model.AuthProvider,
) error {

	query := `
	INSERT INTO auth.auth_providers (
		user_id,
		provider,
		wallet_address
	)
	VALUES (
		$1,
		'wallet',
		$2
	)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		provider.UserID,
		provider.WalletAddress,
	)

	return err
}
