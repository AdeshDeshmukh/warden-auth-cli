package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/config"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var commonPasswords = []string{
	"password", "Password1", "12345678", "qwerty123", "letmein1",
	"welcome1", "admin123", "iloveyou", "sunshine1", "monkey123",
	"dragon12", "master12", "abc12345", "football", "baseball1",
	"shadow12", "michael1", "jessica1", "princess", "superman1",
}

type AuthService struct {
	userRepo  domain.UserRepository
	auditRepo domain.AuditRepository
	cfg       *config.Config
	log       *slog.Logger
}

func NewAuthService(
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
	cfg *config.Config,
	log *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
		log:       log,
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}

	if err := s.CheckPasswordStrength(password); err != nil {
		return err
	}

	_, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("failed to check username availability: %w", err)
	}

	hash, err := hashPassword(password, s.cfg)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	_ = s.auditRepo.Log(ctx, domain.AuditLog{
		UserID:   user.ID,
		Username: username,
		Event:    domain.EventUserRegistered,
		Detail:   `{}`,
	})

	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			dummySalt := make([]byte, 16)
			argon2.IDKey(
				[]byte(password),
				dummySalt,
				s.cfg.ArgonIterations,
				s.cfg.ArgonMemory,
				s.cfg.ArgonParallelism,
				32,
			)
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if user.IsLocked() {
		_ = s.auditRepo.Log(ctx, domain.AuditLog{
			UserID:   user.ID,
			Username: username,
			Event:    domain.EventLoginFailed,
			Detail:   `{"reason":"account_locked"}`,
		})
		return user, domain.ErrAccountLocked
	}

	if user.FailedAttempts >= 2 {
		time.Sleep(2 * time.Second)
	}

	if !verifyPassword(password, user.PasswordHash, s.cfg) {
		newCount := user.FailedAttempts + 1
		var lockedUntil *time.Time

		if newCount >= s.cfg.MaxFailedAttempts {
			t := time.Now().UTC().Add(s.cfg.LockoutDuration)
			lockedUntil = &t

			_ = s.auditRepo.Log(ctx, domain.AuditLog{
				UserID:   user.ID,
				Username: username,
				Event:    domain.EventAccountLocked,
				Detail:   fmt.Sprintf(`{"locked_until":"%s"}`, t.Format(time.RFC3339)),
			})
		}

		_ = s.userRepo.UpdateFailedAttempts(ctx, user.ID, newCount, lockedUntil)

		remaining := s.cfg.MaxFailedAttempts - newCount
		if remaining < 0 {
			remaining = 0
		}

		_ = s.auditRepo.Log(ctx, domain.AuditLog{
			UserID:   user.ID,
			Username: username,
			Event:    domain.EventLoginFailed,
			Detail:   fmt.Sprintf(`{"reason":"invalid_password","attempts_remaining":%d}`, remaining),
		})

		return nil, domain.ErrInvalidCredentials
	}

	_ = s.userRepo.ResetLockout(ctx, user.ID)
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	_ = s.auditRepo.Log(ctx, domain.AuditLog{
		UserID:   user.ID,
		Username: username,
		Event:    domain.EventLoginSuccess,
		Detail:   `{}`,
	})

	user.FailedAttempts = 0
	user.LockedUntil = nil

	return user, nil
}

func (s *AuthService) CheckPasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: must be at least 8 characters", domain.ErrWeakPassword)
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: must contain at least one uppercase letter", domain.ErrWeakPassword)
	}
	if !hasLower {
		return fmt.Errorf("%w: must contain at least one lowercase letter", domain.ErrWeakPassword)
	}
	if !hasDigit {
		return fmt.Errorf("%w: must contain at least one number", domain.ErrWeakPassword)
	}

	lower := strings.ToLower(password)
	for _, blocked := range commonPasswords {
		if lower == strings.ToLower(blocked) {
			return fmt.Errorf("%w: password is too common", domain.ErrWeakPassword)
		}
	}

	return nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func hashPassword(password string, cfg *config.Config) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		cfg.ArgonIterations,
		cfg.ArgonMemory,
		cfg.ArgonParallelism,
		32,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		cfg.ArgonMemory,
		cfg.ArgonIterations,
		cfg.ArgonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func verifyPassword(password, encodedHash string, cfg *config.Config) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		cfg.ArgonIterations,
		cfg.ArgonMemory,
		cfg.ArgonParallelism,
		32,
	)

	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return fmt.Errorf("username must be between 3 and 32 characters")
	}
	for _, ch := range username {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return fmt.Errorf("username may only contain letters, numbers, and underscores")
		}
	}
	return nil
}