package cli

import (
	"context"
	"errors"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
)

func handleWhoAmI(ctx context.Context, a *App, session *domain.Session) {
	user, err := a.auth.GetUserByID(ctx, a.currentUser.ID)
	if err != nil {
		PrintError("Failed to fetch user details.")
		return
	}

	PrintUserProfile(user, session)
}

func handleEnable2FA(ctx context.Context, a *App) {
	user, err := a.auth.GetUserByID(ctx, a.currentUser.ID)
	if err != nil {
		PrintError("Failed to fetch user details.")
		return
	}

	if user.TOTPEnabled {
		PrintError("2FA is already enabled on your account.")
		return
	}

	secret, otpauthURL, err := a.totp.Generate(user.Username)
	if err != nil {
		PrintError("Failed to generate 2FA secret.")
		return
	}

	PrintQRCode(otpauthURL)
	PrintInfo("Manual entry key: " + secret)

	code := Prompt("Enter the 6-digit code from your authenticator app to confirm")
	if code == "" {
		PrintError("Code cannot be empty. 2FA was not enabled.")
		return
	}

	err = a.totp.Enable(ctx, user, secret, code)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTOTPCode) {
			PrintError("Invalid code. 2FA was not enabled. Please try again.")
			return
		}
		if errors.Is(err, domain.ErrTOTPAlreadyEnabled) {
			PrintError("2FA is already enabled.")
			return
		}
		PrintError("Failed to enable 2FA. Please try again.")
		return
	}

	a.currentUser.TOTPEnabled = true
	PrintSuccess("2FA enabled successfully. Your account is now protected.")
}

func handleDisable2FA(ctx context.Context, a *App) {
	user, err := a.auth.GetUserByID(ctx, a.currentUser.ID)
	if err != nil {
		PrintError("Failed to fetch user details.")
		return
	}

	if !user.TOTPEnabled {
		PrintError("2FA is not enabled on your account.")
		return
	}

	code := Prompt("Enter your current 6-digit 2FA code to confirm")
	if code == "" {
		PrintError("Code cannot be empty. 2FA was not disabled.")
		return
	}

	err = a.totp.Disable(ctx, user, code)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTOTPCode) {
			PrintError("Invalid code. 2FA was not disabled.")
			return
		}
		if errors.Is(err, domain.ErrTOTPNotEnabled) {
			PrintError("2FA is not enabled.")
			return
		}
		PrintError("Failed to disable 2FA. Please try again.")
		return
	}

	a.currentUser.TOTPEnabled = false
	PrintSuccess("2FA disabled successfully.")
}

func handleLogout(ctx context.Context, a *App) {
	err := a.sessions.Invalidate(
		ctx,
		a.currentToken,
		a.currentUser.Username,
		a.currentUser.ID,
	)
	if err != nil {
		PrintError("Failed to logout cleanly.")
	}

	a.currentToken = ""
	a.currentUser = nil

	PrintSuccess("Logged out successfully. Goodbye.")
}