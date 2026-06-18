package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateFailedAttempts(ctx context.Context, id string, count int, lockedUntil *time.Time) error
	UpdateLastLogin(ctx context.Context, id string) error
	UpdateTOTP(ctx context.Context, id string, secret string, enabled bool) error
	ResetLockout(ctx context.Context, id string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type AuditRepository interface {
	Log(ctx context.Context, entry AuditLog) error
}