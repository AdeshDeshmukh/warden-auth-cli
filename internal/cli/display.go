package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

func PrintBanner() {
	pterm.Println()
	bannerStyle := pterm.NewStyle(pterm.FgCyan, pterm.Bold)
	bannerStyle.Println(`  ██╗    ██╗ █████╗ ██████╗ ██████╗ ███████╗███╗   ██╗`)
	bannerStyle.Println(`  ██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔════╝████╗  ██║`)
	bannerStyle.Println(`  ██║ █╗ ██║███████║██████╔╝██║  ██║█████╗  ██╔██╗ ██║`)
	bannerStyle.Println(`  ██║███╗██║██╔══██║██╔══██╗██║  ██║██╔══╝  ██║╚██╗██║`)
	bannerStyle.Println(`  ╚███╔███╔╝██║  ██║██║  ██║██████╔╝███████╗██║ ╚████║`)
	bannerStyle.Println(`   ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═══╝`)
	pterm.Println()
	pterm.FgGray.Println("  Warden Auth CLI  v1.0.0  —  Secure. Audited. Containerized.")
	pterm.Println()
	pterm.FgYellow.Println("  Type 'help' to see available commands.")
	pterm.Println()
}

func PrintSuccess(msg string) {
	pterm.Success.Println(msg)
}

func PrintError(msg string) {
	pterm.Error.Println(msg)
}

func PrintWarning(msg string) {
	pterm.Warning.Println(msg)
}

func PrintInfo(msg string) {
	pterm.Info.Println(msg)
}

func PrintUserProfile(user *domain.User, session *domain.Session) {
	pterm.Println()

	tableData := pterm.TableData{
		{"Field", "Value"},
		{"Username", user.Username},
		{"Registered", user.CreatedAt.Format("2006-01-02")},
		{"Last Login", formatLastLogin(user.LastLoginAt)},
		{"2FA Status", format2FAStatus(user.TOTPEnabled)},
		{"Session Expires", "in " + FormatDuration(session.TimeRemaining())},
		{"Account Status", formatAccountStatus(user)},
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData).
		Render()

	pterm.Println()
}

func PrintHelp(loggedIn bool) {
	pterm.Println()

	var tableData pterm.TableData

	if loggedIn {
		tableData = pterm.TableData{
			{"Command", "Description"},
			{"whoami", "Show your profile and session details"},
			{"enable-2fa", "Set up TOTP two-factor authentication"},
			{"disable-2fa", "Remove two-factor authentication"},
			{"logout", "End your current session"},
			{"help", "Show available commands"},
		}
	} else {
		tableData = pterm.TableData{
			{"Command", "Description"},
			{"register", "Create a new account"},
			{"login", "Sign in to your account"},
			{"help", "Show available commands"},
			{"exit", "Quit the application"},
		}
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData).
		Render()

	pterm.Println()
}

func PrintQRCode(otpauthURL string) {
	pterm.Println()
	pterm.Info.Println("Scan the QR code below with Google Authenticator or Authy:")
	pterm.Println()
	pterm.FgCyan.Println("  otpauth URL: " + otpauthURL)
	pterm.Println()
}

func PrintLockoutWarning(remaining time.Duration) {
	pterm.Println()
	pterm.Error.Println("Too many failed attempts.")
	pterm.Warning.Printf("Account locked. Try again in %s.\n", FormatDuration(remaining))
	pterm.Println()
}

func PrintSessionWarning(remaining time.Duration) {
	pterm.Warning.Printf("Session expires in %s.\n", FormatDuration(remaining))
}

func Prompt(label string) string {
	pterm.FgCyan.Printf("  %s: ", label)
	var input string
	fmt.Scanln(&input)
	return input
}

func PromptPassword(label string) string {
	pterm.FgCyan.Printf("  %s: ", label)

	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return ""
	}

	return string(bytePassword)
}

func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func formatLastLogin(t *time.Time) string {
	if t == nil {
		return "Never"
	}

	diff := time.Since(*t)

	if diff < time.Minute {
		return "Just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d minute(s) ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d hour(s) ago", int(diff.Hours()))
	}

	return t.Format("2006-01-02 15:04")
}

func format2FAStatus(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func formatAccountStatus(user *domain.User) string {
	if user.IsLocked() {
		return "Locked"
	}
	return "Active"
}