// Package repository handles all database queries for auth-service.
// Rule: ONLY SQL lives here. No business logic. No HTTP stuff.
//
// Two repositories:
//   - UserRepository     → queries against the `users` table
//   - TokenRepository    → queries against `refresh_tokens` + Redis

package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preloved-services/auth-service/internal/model"
)

// UserRepository handles all queries on the `users` table.
type UserRepo struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a UserRepository.
// Call this once in main.go and pass it to AuthService.
func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser implements the CreateUser method.
// Create a new user in the database.
// Parameters:
// - ctx: The context for the database operation
// - user: The user to create
// Returns:
// - error: An error if the user could not be created
func (r *UserRepo) Create(
	ctx context.Context,
	user *model.User,
) (*model.User, error) {

	query := `
	INSERT INTO auth.users
	(
		email,
		username,
		role,
		status
	)
	VALUES
	(
		$1,
		$2,
		$3,
		$4
	)
	RETURNING
		id,
		email,
		username,
		role,
		status,
		avatar_url,
		is_verified,
		created_at,
		updated_at
	`

	var created model.User

	err := r.db.QueryRow(
		ctx,
		query,
		user.Email,
		user.Username,
		user.Role,
		user.Status,
	).Scan(
		&created.ID,
		&created.Email,
		&created.Username,
		&created.Role,
		&created.Status,
		&created.AvatarURL,
		&created.IsVerified,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			switch pgErr.ConstraintName {

			case "users_email_key":
				return nil, ErrDuplicateEmail

			case "users_username_key":
				return nil, ErrDuplicateUsername
			}
		}

		return nil, err
	}
	return &created, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, username, role, status, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var u model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.Role,
		&u.Status,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) GetByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {

	query := `
	SELECT
		id,
		email,
		username,
		role,
		status,
		avatar_url,
		is_verified,
		deleted_at,
		created_at,
		updated_at
	FROM auth.users
	WHERE email = $1
	AND deleted_at IS NULL
	LIMIT 1
	`

	var user model.User

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Role,
		&user.Status,
		&user.AvatarURL,
		&user.IsVerified,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) EmailExists(
	ctx context.Context,
	email string,
) (bool, error) {

	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1
			FROM auth.users
			WHERE email = $1
			AND deleted_at IS NULL
		)
		`,
		email,
	).Scan(&exists)

	return exists, err
}

func (r *UserRepo) UsernameExists(
	ctx context.Context,
	username string,
) (bool, error) {

	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1
			FROM auth.users
			WHERE username = $1
			AND deleted_at IS NULL
		)
		`,
		username,
	).Scan(&exists)

	return exists, err
}
func (r *UserRepo) VerifyUser(
	ctx context.Context,
	userID string,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE auth.users
		SET
			is_verified = TRUE,
			status = 'active'
		WHERE id = $1
		`,
		userID,
	)

	return err
}

func (r *UserRepo) SoftDelete(
	ctx context.Context,
	userID string,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE auth.users
		SET deleted_at = NOW()
		WHERE id = $1
		`,
		userID,
	)

	return err
}
