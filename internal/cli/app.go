package cli

import (
	"context"
	"io"
	"strings"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
	"github.com/AdeshDeshmukh/warden-auth-cli/internal/service"
	"github.com/chzyer/readline"
	"github.com/pterm/pterm"
)

type App struct {
	rl           *readline.Instance
	auth         *service.AuthService
	sessions     *service.SessionService
	totp         *service.TOTPService
	currentToken string
	currentUser  *domain.User
}

func NewApp(
	auth *service.AuthService,
	sessions *service.SessionService,
	totp *service.TOTPService,
) (*App, error) {
	app := &App{
		auth:     auth,
		sessions: sessions,
		totp:     totp,
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "warden ❯ ",
		HistoryLimit:    100,
		AutoComplete:    readline.NewPrefixCompleter(buildPreLoginCompleters()...),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}

	app.rl = rl
	return app, nil
}

func (a *App) Run(ctx context.Context) {
	defer a.rl.Close()

	PrintBanner()

	for {
		a.updatePrompt()

		line, err := a.rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		}
		if err == io.EOF {
			break
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		a.dispatch(ctx, command)
	}

	pterm.Println()
	PrintInfo("Goodbye.")
}

func (a *App) dispatch(ctx context.Context, command string) {
	if a.currentToken == "" {
		a.dispatchPreLogin(ctx, command)
		return
	}

	session, err := a.sessions.Validate(ctx, a.currentToken)
	if err != nil {
		PrintWarning("Your session has expired. Please login again.")
		a.currentToken = ""
		a.currentUser = nil
		a.updatePrompt()
		return
	}

	if session.TimeRemaining().Minutes() < 5 {
		PrintSessionWarning(session.TimeRemaining())
	}

	a.dispatchPostLogin(ctx, command, session)
}

func (a *App) dispatchPreLogin(ctx context.Context, command string) {
	switch command {
	case "register":
		handleRegister(ctx, a)
	case "login":
		handleLogin(ctx, a)
	case "help":
		PrintHelp(false)
	case "exit":
		PrintInfo("Goodbye.")
		a.rl.Close()
	default:
		PrintError("Unknown command '" + command + "'. Type 'help' to see available commands.")
	}
}

func (a *App) dispatchPostLogin(ctx context.Context, command string, session *domain.Session) {
	switch command {
	case "whoami":
		handleWhoAmI(ctx, a, session)
	case "enable-2fa":
		handleEnable2FA(ctx, a)
	case "disable-2fa":
		handleDisable2FA(ctx, a)
	case "logout":
		handleLogout(ctx, a)
	case "help":
		PrintHelp(true)
	default:
		PrintError("Unknown command '" + command + "'. Type 'help' to see available commands.")
	}
}

func (a *App) updatePrompt() {
	if a.currentUser == nil {
		a.rl.SetPrompt("warden ❯ ")
		a.rl.Config.AutoComplete = readline.NewPrefixCompleter(buildPreLoginCompleters()...)
	} else {
		a.rl.SetPrompt("warden [" + a.currentUser.Username + "] ❯ ")
		a.rl.Config.AutoComplete = readline.NewPrefixCompleter(buildPostLoginCompleters()...)
	}
}

func buildPreLoginCompleters() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("register"),
		readline.PcItem("login"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	}
}

func buildPostLoginCompleters() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{
		readline.PcItem("whoami"),
		readline.PcItem("enable-2fa"),
		readline.PcItem("disable-2fa"),
		readline.PcItem("logout"),
		readline.PcItem("help"),
	}
}