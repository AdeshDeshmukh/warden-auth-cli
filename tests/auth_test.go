package tests

import (
	"testing"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/config"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/service"
)

func TestPasswordStrength_TooShort(t *testing.T) {
	svc := newTestAuthService()
	err := svc.CheckPasswordStrength("Ab1")
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
}

func TestPasswordStrength_NoUppercase(t *testing.T) {
	svc := newTestAuthService()
	err := svc.CheckPasswordStrength("abcdefg1")
	if err == nil {
		t.Fatal("expected error for missing uppercase, got nil")
	}
}

func TestPasswordStrength_NoLowercase(t *testing.T) {
	svc := newTestAuthService()
	err := svc.CheckPasswordStrength("ABCDEFG1")
	if err == nil {
		t.Fatal("expected error for missing lowercase, got nil")
	}
}

func TestPasswordStrength_NoDigit(t *testing.T) {
	svc := newTestAuthService()
	err := svc.CheckPasswordStrength("Abcdefgh")
	if err == nil {
		t.Fatal("expected error for missing digit, got nil")
	}
}

func TestPasswordStrength_CommonPassword(t *testing.T) {
	svc := newTestAuthService()
	err := svc.CheckPasswordStrength("Password1")
	if err == nil {
		t.Fatal("expected error for common password, got nil")
	}
}

func TestPasswordStrength_ValidPassword(t *testing.T) {
	svc := newTestAuthService()
	err := svc.CheckPasswordStrength("Warden@2025")
	if err != nil {
		t.Fatalf("expected no error for valid password, got: %v", err)
	}
}

func newTestAuthService() *service.AuthService {
	cfg := config.Load()
	return service.NewAuthService(nil, nil, cfg, nil)
}