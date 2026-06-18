package cli

import (
	"context"
	"errors"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
)

func handleRegister(ctx context.Context, a *App) {
	username := Prompt("Username")
	if username == "" {
		PrintError("Username cannot be empty.")
		return
	}

	password := PromptPassword("Password")
	if password == "" {
		PrintError("Password cannot be empty.")
		return
	}

	confirm := PromptPassword("Confirm Password")
	if password != confirm {
		PrintError("Passwords do not match.")
		return
	}

	err := a.auth.Register(ctx, username, password)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			PrintError("Username is already taken. Please choose another.")
			return
		}
		if errors.Is(err, domain.ErrWeakPassword) {
			PrintError(err.Error())
			return
		}
		PrintError("Registration failed. Please try again.")
		return
	}

	PrintSuccess("Account created successfully. You can now login.")
}

func handleLogin(ctx context.Context, a *App) {
	username := Prompt("Username")
	if username == "" {
		PrintError("Username cannot be empty.")
		return
	}

	password := PromptPassword("Password")
	if password == "" {
		PrintError("Password cannot be empty.")
		return
	}

	user, err := a.auth.Login(ctx, username, password)
	if err != nil {
		if errors.Is(err, domain.ErrAccountLocked) && user != nil {
			PrintLockoutWarning(user.LockoutRemaining())
			return
		}
		if errors.Is(err, domain.ErrInvalidCredentials) {
			PrintError("Invalid credentials. Please try again.")
			return
		}
		PrintError("Login failed. Please try again.")
		return
	}

	if user.TOTPEnabled {
		code := Prompt("2FA Code")
		if code == "" {
			PrintError("2FA code cannot be empty.")
			return
		}

		if !a.totp.Verify(user.TOTPSecret, code) {
			PrintError("Invalid 2FA code.")
			return
		}
	}

	rawToken, err := a.sessions.Create(ctx, user)
	if err != nil {
		PrintError("Failed to create session. Please try again.")
		return
	}

	a.currentToken = rawToken
	a.currentUser = user

	PrintSuccess("Welcome back, " + user.Username + "!")

	session, err := a.sessions.Validate(ctx, rawToken)
	if err == nil {
		PrintUserProfile(user, session)
	}
}