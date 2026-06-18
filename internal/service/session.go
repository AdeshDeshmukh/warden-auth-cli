package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/config"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
	"github.com/google/uuid"
)

type SessionService struct {
	sessionRepo domain.SessionRepository
	auditRepo   domain.AuditRepository
	cfg         *config.Config
	log         *slog.Logger
}

func NewSessionService(
	sessionRepo domain.SessionRepository,
	auditRepo domain.AuditRepository,
	cfg *config.Config,
	log *slog.Logger,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		cfg:         cfg,
		log:         log,
	}
}

func (s *SessionService) Create(ctx context.Context, user *domain.User) (string, error) {
	if err := s.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return "", fmt.Errorf("failed to clear existing sessions: %w", err)
	}

	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	rawToken := hex.EncodeToString(buf)
	tokenHash := hashToken(rawToken)

	session := &domain.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.cfg.SessionTimeout),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	_ = s.auditRepo.Log(ctx, domain.AuditLog{
		UserID:   user.ID,
		Username: user.Username,
		Event:    domain.EventSessionCreated,
		Detail:   fmt.Sprintf(`{"expires_at":"%s"}`, session.ExpiresAt.Format(time.RFC3339)),
	})

	return rawToken, nil
}

func (s *SessionService) Validate(ctx context.Context, rawToken string) (*domain.Session, error) {
	tokenHash := hashToken(rawToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to fetch session: %w", err)
	}

	if session.IsExpired() {
		_ = s.sessionRepo.DeleteByUserID(ctx, session.UserID)
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

func (s *SessionService) Invalidate(ctx context.Context, rawToken, username, userID string) error {
	tokenHash := hashToken(rawToken)

	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, session.UserID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	_ = s.auditRepo.Log(ctx, domain.AuditLog{
		UserID:   userID,
		Username: username,
		Event:    domain.EventLogout,
		Detail:   `{}`,
	})

	return nil
}

func (s *SessionService) CleanExpired(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}

func hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}