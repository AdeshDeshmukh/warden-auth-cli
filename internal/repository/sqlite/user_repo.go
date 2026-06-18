package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, username, password_hash, totp_secret, totp_enabled,
		                   failed_attempts, locked_until, last_login_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.TOTPSecret,
		boolToInt(user.TOTPEnabled),
		user.FailedAttempts,
		user.LockedUntil,
		user.LastLoginAt,
		user.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT id, username, password_hash, totp_secret, totp_enabled,
		       failed_attempts, locked_until, last_login_at, created_at
		FROM users WHERE username = ?`

	return r.scanUser(r.db.QueryRowContext(ctx, query, username))
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT id, username, password_hash, totp_secret, totp_enabled,
		       failed_attempts, locked_until, last_login_at, created_at
		FROM users WHERE id = ?`

	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *UserRepo) UpdateFailedAttempts(ctx context.Context, id string, count int, lockedUntil *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE users SET failed_attempts = ?, locked_until = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, count, lockedUntil, id)
	if err != nil {
		return fmt.Errorf("failed to update failed attempts: %w", err)
	}

	return nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE users SET last_login_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *UserRepo) UpdateTOTP(ctx context.Context, id string, secret string, enabled bool) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE users SET totp_secret = ?, totp_enabled = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, secret, boolToInt(enabled), id)
	if err != nil {
		return fmt.Errorf("failed to update TOTP: %w", err)
	}

	return nil
}

func (r *UserRepo) ResetLockout(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reset lockout: %w", err)
	}

	return nil
}

func (r *UserRepo) scanUser(row *sql.Row) (*domain.User, error) {
	var user domain.User
	var totpEnabled int
	var lockedUntil, lastLoginAt sql.NullTime
	var totpSecret sql.NullString

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&totpSecret,
		&totpEnabled,
		&user.FailedAttempts,
		&lockedUntil,
		&lastLoginAt,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	user.TOTPEnabled = totpEnabled == 1
	user.TOTPSecret = totpSecret.String

	if lockedUntil.Valid {
		t := lockedUntil.Time
		user.LockedUntil = &t
	}

	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		user.LastLoginAt = &t
	}

	return &user, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}