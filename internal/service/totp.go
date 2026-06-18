package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
	"github.com/pquerna/otp/totp"
)

type TOTPService struct {
	userRepo  domain.UserRepository
	auditRepo domain.AuditRepository
	log       *slog.Logger
}

func NewTOTPService(
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
	log *slog.Logger,
) *TOTPService {
	return &TOTPService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		log:       log,
	}
}

func (s *TOTPService) Generate(username string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Warden Auth",
		AccountName: username,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

func (s *TOTPService) Verify(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (s *TOTPService) Enable(ctx context.Context, user *domain.User, secret, code string) error {
	if user.TOTPEnabled {
		return domain.ErrTOTPAlreadyEnabled
	}

	if !s.Verify(secret, code) {
		return domain.ErrInvalidTOTPCode
	}

	if err := s.userRepo.UpdateTOTP(ctx, user.ID, secret, true); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	_ = s.auditRepo.Log(ctx, domain.AuditLog{
		UserID:   user.ID,
		Username: user.Username,
		Event:    domain.EventTOTPEnabled,
		Detail:   `{}`,
	})

	return nil
}

func (s *TOTPService) Disable(ctx context.Context, user *domain.User, code string) error {
	if !user.TOTPEnabled {
		return domain.ErrTOTPNotEnabled
	}

	if !s.Verify(user.TOTPSecret, code) {
		return domain.ErrInvalidTOTPCode
	}

	if err := s.userRepo.UpdateTOTP(ctx, user.ID, "", false); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	_ = s.auditRepo.Log(ctx, domain.AuditLog{
		UserID:   user.ID,
		Username: user.Username,
		Event:    domain.EventTOTPDisabled,
		Detail:   `{}`,
	})

	return nil
}