package tests

import (
	"testing"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/service"
)

func TestTOTPGenerate_ReturnsSecretAndURL(t *testing.T) {
	svc := newTestTOTPService()

	secret, url, err := svc.Generate("adesh")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if secret == "" {
		t.Fatal("expected non-empty secret")
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestTOTPGenerate_DifferentSecretsEachTime(t *testing.T) {
	svc := newTestTOTPService()

	secret1, _, err := svc.Generate("adesh")
	if err != nil {
		t.Fatalf("first generate failed: %v", err)
	}

	secret2, _, err := svc.Generate("adesh")
	if err != nil {
		t.Fatalf("second generate failed: %v", err)
	}

	if secret1 == secret2 {
		t.Fatal("expected different secrets on each generate call")
	}
}

func TestTOTPVerify_InvalidCode(t *testing.T) {
	svc := newTestTOTPService()

	secret, _, err := svc.Generate("adesh")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	valid := svc.Verify(secret, "000000")
	if valid {
		t.Fatal("expected invalid code to return false")
	}
}

func TestTOTPVerify_EmptyCode(t *testing.T) {
	svc := newTestTOTPService()

	secret, _, err := svc.Generate("adesh")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	valid := svc.Verify(secret, "")
	if valid {
		t.Fatal("expected empty code to return false")
	}
}

func newTestTOTPService() *service.TOTPService {
	return service.NewTOTPService(nil, nil, nil)
}