# 🔐 Warden Auth CLI

<div align="center">

<pre style="text-align:center; font-family: monospace;">
  ██╗    ██╗ █████╗ ██████╗ ██████╗ ███████╗███╗   ██╗
  ██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔════╝████╗  ██║
  ██║ █╗ ██║███████║██████╔╝██║  ██║█████╗  ██╔██╗ ██║
  ██║███╗██║██╔══██║██╔══██╗██║  ██║██╔══╝  ██║╚██╗██║
  ╚███╔███╔╝██║  ██║██║  ██║██████╔╝███████╗██║ ╚████║
   ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═══╝
</pre>

**Production-grade containerized CLI authentication system**

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Distroless-2496ED?style=flat&logo=docker)](https://docker.com)
[![SQLite](https://img.shields.io/badge/SQLite-WAL_Mode-003B57?style=flat&logo=sqlite)](https://sqlite.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Argon2id-red?style=flat)](https://en.wikipedia.org/wiki/Argon2)

</div>

---

## 📌 Overview

**Warden Auth CLI** is a secure, containerized command-line authentication system built entirely in Go. It implements user registration, password-based authentication, optional TOTP two-factor authentication, and persistent session management — all running inside Docker with SQLite for storage.

This project was built to demonstrate production-grade backend engineering practices including clean architecture, layered separation of concerns, security-first design, and professional DevOps packaging.

---

## ✨ Key Highlights

| Feature | Implementation |
|---|---|
| Password Hashing | **Argon2id** — OWASP 2023 recommended, memory-hard algorithm |
| Session Security | **SHA256-hashed tokens** — raw token never stored on disk |
| Account Protection | **Progressive lockout** — escalating delays before hard lock |
| Two-Factor Auth | **TOTP** — Google Authenticator and Authy compatible |
| Audit Trail | **Structured JSON logs** — every security event recorded |
| Container Security | **Distroless image** — zero shell, zero OS attack surface |
| Architecture | **5-layer clean architecture** — fully decoupled and testable |
| Data Persistence | **SQLite with WAL mode** — survives container restarts |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Docker Container                         │
│                   gcr.io/distroless/static                      │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Warden Auth CLI                       │   │
│  │                                                          │   │
│  │  ┌───────────┐   ┌───────────┐   ┌──────────────────┐    │   │
│  │  │    CLI    │──▶│  Service  │──▶│   Repository     │    │   │
│  │  │   Layer   │   │   Layer   │   │     Layer        │    │   │
│  │  │           │   │           │   │                  │    │   │
│  │  │ readline  │   │  Argon2id │   │  SQLite + WAL    │    │   │
│  │  │ pterm     │   │  TOTP     │   │  Migrations      │    │   │
│  │  │ display   │   │  Sessions │   │  Audit Logs      │    │   │
│  │  └───────────┘   └───────────┘   └──────────────────┘    │   │
│  │         │               │                  │             │   │
│  │         ▼               ▼                  ▼             │   │
│  │  ┌──────────────────────────────────────────────────┐    │   │
│  │  │                  Domain Layer                    │    │   │
│  │  │     Models • Interfaces • Sentinel Errors        │    │   │
│  │  └──────────────────────────────────────────────────┘    │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                  │
│              ┌───────────────┴───────────────┐                  │
│              ▼                               ▼                  │
│     ┌─────────────────┐           ┌─────────────────┐           │
│     │   db_data vol   │           │  log_data vol   │           │
│     │   warden.db     │           │  warden.log     │           │
│     │   (SQLite)      │           │  (JSON audit)   │           │
│     └─────────────────┘           └─────────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Package | Rule |
|---|---|---|
| **Domain** | `internal/domain` | Zero external imports. Defines models, interfaces, errors. |
| **Repository** | `internal/repository/sqlite` | All SQL lives here. Maps DB errors to domain errors. |
| **Service** | `internal/service` | Business logic only. No SQL, no CLI, no display code. |
| **CLI** | `internal/cli` | Input, routing, display only. No business logic. |
| **Config** | `internal/config` | ENV loading with defaults. Single Config struct. |
| **Logger** | `internal/logger` | Dual slog output — JSON to file, text to stdout. |

---

## 🚀 Quick Start

**Requirements:** Docker and Docker Compose installed.

```bash
git clone https://github.com/AdeshDeshmukh/warden-auth-cli.git
cd warden-auth-cli
cp .env.example .env
docker-compose run --rm warden
```

That is it. The application starts, migrations run automatically, and you are at the prompt.

---

## 📖 Commands Reference

### Before Login

| Command | Description |
|---|---|
| `register` | Create a new user account with password validation |
| `login` | Authenticate — prompts for TOTP code if 2FA is enabled |
| `help` | Display all available commands |
| `exit` | Quit the application cleanly |

### After Login

| Command | Description |
|---|---|
| `whoami` | Display full profile — username, registration date, 2FA status, session expiry, last login |
| `enable-2fa` | Generate TOTP secret, display setup key, verify before saving |
| `disable-2fa` | Verify current TOTP code then remove 2FA from account |
| `logout` | Invalidate session token and return to login prompt |
| `help` | Display all available commands |

---

## 🖥️ Usage Examples

### Register

```
warden ❯ register
  Username: adesh
  Password: ••••••••
  Confirm Password: ••••••••
✅ Account created successfully. You can now login.
```

### Login

```
warden ❯ login
  Username: adesh
  Password: ••••••••
✅ Welcome back, adesh!

┌──────────────────────────────────────────────┐
│ Field           │ Value                      │
├─────────────────┼────────────────────────────┤
│ Username        │ adesh                      │
│ Registered      │ 2025-06-18                 │
│ Last Login      │ Just now                   │
│ 2FA Status      │ Disabled                   │
│ Session Expires │ in 30m 0s                  │
│ Account Status  │ Active                     │
└──────────────────────────────────────────────┘
```

### Login with 2FA Enabled

```
warden ❯ login
  Username: adesh
  Password: ••••••••
  2FA Code: ••••••
✅ Welcome back, adesh!
```

### Enable Two-Factor Authentication

```
warden [adesh] ❯ enable-2fa
ℹ️  Scan the QR code below with Google Authenticator or Authy:
ℹ️  otpauth URL: otpauth://totp/Warden%20Auth:adesh?...
ℹ️  Manual entry key: JBSWY3DPEHPK3PXP
  Enter the 6-digit code from your authenticator app to confirm: ••••••
✅ 2FA enabled successfully. Your account is now protected.
```

### Failed Login with Progressive Lockout

```
warden ❯ login
  Username: adesh
  Password: ••••••••
❌ Invalid credentials. Please try again.

warden ❯ login
  Username: adesh
  Password: ••••••••
❌ Invalid credentials. Please try again.

[after 5 failed attempts]

❌ Too many failed attempts.
⚠️  Account locked. Try again in 15m 0s.
```

---

## ⚙️ Configuration

Copy `.env.example` to `.env` and modify as needed. All values below are defaults.

| Variable | Default | Description |
|---|---|---|
| `DB_PATH` | `./data/warden.db` | SQLite database file location |
| `LOG_PATH` | `./logs/warden.log` | JSON audit log file location |
| `SESSION_TIMEOUT` | `30m` | Session duration before automatic expiry |
| `MAX_FAILED_ATTEMPTS` | `5` | Failed login threshold before account lockout |
| `LOCKOUT_DURATION` | `15m` | Duration of account lockout after threshold reached |
| `ARGON_MEMORY` | `65536` | Argon2id memory parameter in KB (64MB) |
| `ARGON_ITERATIONS` | `3` | Argon2id iteration count |
| `ARGON_PARALLELISM` | `2` | Argon2id parallel thread count |

---

## 🔒 Security Design Decisions

### Argon2id over bcrypt

bcrypt is CPU-hard only. Argon2id is both CPU-hard and memory-hard. The memory requirement (64MB per attempt) defeats GPU-based brute force attacks because GPU cores share limited memory bandwidth — thousands of parallel attempts become impossible. OWASP has recommended Argon2id as the preferred password hashing algorithm since 2023. The parameters used here meet OWASP minimum requirements.

### SHA256 session token hashing

Storing raw session tokens in a database is equivalent to storing plaintext passwords. If the database file is ever compromised, an attacker would immediately control all active sessions. By storing `SHA256(rawToken)` instead, the raw token exists only in the running process memory and is never written to persistent storage. SHA256 is a one-way function — the stored hash cannot be reversed.

### Progressive lockout over hard cutoff

A hard cutoff at N attempts is easy to detect and probe during an attack. Progressive delays (2 second artificial sleep from the third failed attempt onward) make brute force attempts slow and painful before the hard lockout triggers at 5 attempts. The exact countdown timer shown on lockout informs legitimate users without leaking timing internals to attackers.

### Single session per user

When a new login succeeds, all existing sessions for that user are deleted before the new session is created. This prevents session fixation attacks where an attacker plants a known session identifier and waits for the victim to authenticate with it. It also prevents stale sessions from accumulating indefinitely.

### Distroless final Docker image

Standard base images (ubuntu, alpine) contain shells, package managers, and hundreds of OS-level binaries. Each is a potential attack vector. The `gcr.io/distroless/static-debian12` image contains only the compiled binary and minimal runtime. There is no `/bin/sh` — shell injection is structurally impossible. There is no package manager — supply chain attacks via OS packages cannot occur.

### Pure Go SQLite driver

Using `modernc.org/sqlite` instead of the CGo-based `mattn/go-sqlite3` means `CGO_ENABLED=0` during build. This produces a fully static binary with zero shared library dependencies, which is what makes the distroless final image possible. A CGo build would require glibc in the final image, forcing a larger and less secure base.

### Timing attack prevention

When a username does not exist, the login handler still runs the full Argon2id computation on a dummy value before returning an error. Without this, an attacker could enumerate valid usernames by measuring response time — a non-existent user would return faster than an existing one requiring hash verification.

### TOTP code verified before secret is saved

When enabling 2FA, the user must successfully enter a valid six-digit code before the secret is written to the database. This proves the user has correctly scanned the setup key and their authenticator app is generating valid codes. Without this verification step, a user could enable 2FA with a misconfigured app and then be permanently unable to log in.

### Audit logs as a first-class feature

Every authentication event writes simultaneously to the `audit_logs` SQLite table and to a structured JSON log file. The JSON format is directly ingestible by SIEM tools such as Splunk, Elastic Stack, and Datadog. The `username` field is denormalized intentionally — the audit trail remains intact even if a user account is deleted.

---

## 🗄️ Database Schema

### users

| Column | Type | Description |
|---|---|---|
| `id` | TEXT | UUID primary key — prevents enumeration attacks |
| `username` | TEXT | Unique, 3-32 characters, alphanumeric and underscore |
| `password_hash` | TEXT | Argon2id encoded hash — self-contained with salt and params |
| `totp_secret` | TEXT | TOTP secret key, null when 2FA is disabled |
| `totp_enabled` | INTEGER | 1 when 2FA is active, 0 otherwise |
| `failed_attempts` | INTEGER | Failed login counter — persists across container restarts |
| `locked_until` | DATETIME | Lockout expiry timestamp, null when account is active |
| `last_login_at` | DATETIME | Timestamp of most recent successful authentication |
| `created_at` | DATETIME | Account creation timestamp in UTC |

### sessions

| Column | Type | Description |
|---|---|---|
| `id` | TEXT | UUID primary key |
| `user_id` | TEXT | Foreign key to users.id — cascades on delete |
| `token_hash` | TEXT | SHA256 of raw session token — raw token never stored |
| `expires_at` | DATETIME | Session expiry — validated on every command |
| `created_at` | DATETIME | Session creation timestamp |

### audit_logs

| Column | Type | Description |
|---|---|---|
| `id` | INTEGER | Auto-increment primary key |
| `user_id` | TEXT | Foreign key to users.id — set null on delete |
| `username` | TEXT | Denormalized — audit trail survives account deletion |
| `event` | TEXT | Event type constant from predefined set |
| `detail` | TEXT | JSON metadata string with event-specific context |
| `created_at` | DATETIME | Event timestamp in UTC |

---

## 📊 Audit Events Reference

| Event | Triggered When | Detail Fields |
|---|---|---|
| `USER_REGISTERED` | New account created | `{}` |
| `LOGIN_SUCCESS` | Password verified, session created | `{}` |
| `LOGIN_FAILED` | Wrong password or locked account | `reason`, `attempts_remaining` |
| `ACCOUNT_LOCKED` | Failed attempts hit maximum | `locked_until` |
| `TOTP_ENABLED` | 2FA setup completed and verified | `{}` |
| `TOTP_DISABLED` | 2FA removed from account | `{}` |
| `SESSION_CREATED` | New session token issued | `expires_at` |
| `SESSION_EXPIRED` | Expired session detected on validation | `{}` |
| `LOGOUT` | User explicitly ended session | `{}` |

### Sample JSON Audit Log Entry

```json
{
  "time": "2025-06-18T10:23:01.234Z",
  "level": "INFO",
  "msg": "security_event",
  "event": "LOGIN_SUCCESS",
  "username": "adesh",
  "detail": {}
}
```

---

## 🐳 Docker Details

### Multi-Stage Build

**Stage 1 — builder** uses `golang:1.26-alpine`:
- Downloads all Go module dependencies
- Compiles with `CGO_ENABLED=0` for a fully static binary
- Strips debug info with `-ldflags="-w -s"` for smaller size

**Stage 2 — final** uses `gcr.io/distroless/static-debian12`:
- Contains only the compiled binary
- No shell, no package manager, no OS utilities
- Minimal CVE exposure from base image

### Image Size Comparison

| Base Image | Typical Size |
|---|---|
| ubuntu | ~180MB |
| alpine | ~20MB |
| distroless/static | ~15MB |

### Persistent Volumes

| Volume | Mount Point | Contents |
|---|---|---|
| `db_data` | `/app/data` | SQLite database file |
| `log_data` | `/app/logs` | Structured JSON audit log files |

Both volumes survive container restarts, rebuilds, and updates. Your data is never lost when stopping or updating the container.

---

## 🧪 Tests

```bash
go test -v -race ./tests/...
```

| Test | What It Validates |
|---|---|
| `TestPasswordStrength_TooShort` | Rejects passwords under 8 characters |
| `TestPasswordStrength_NoUppercase` | Rejects passwords without uppercase |
| `TestPasswordStrength_NoLowercase` | Rejects passwords without lowercase |
| `TestPasswordStrength_NoDigit` | Rejects passwords without a digit |
| `TestPasswordStrength_CommonPassword` | Rejects passwords on common blocklist |
| `TestPasswordStrength_ValidPassword` | Accepts a strong valid password |
| `TestTOTPGenerate_ReturnsSecretAndURL` | TOTP generation returns non-empty values |
| `TestTOTPGenerate_DifferentSecretsEachTime` | Each call generates a unique secret |
| `TestTOTPVerify_InvalidCode` | Rejects wrong 6-digit codes |
| `TestTOTPVerify_EmptyCode` | Rejects empty code input |

---

## 📁 Project Structure

```
warden-auth-cli/
├── cmd/
│   └── main.go                     Entry point — config, DB, wiring, shutdown
├── internal/
│   ├── config/
│   │   └── config.go               ENV variable loading with typed defaults
│   ├── domain/
│   │   ├── user.go                 User model with IsLocked, LockoutRemaining
│   │   ├── session.go              Session model with IsExpired, TimeRemaining
│   │   ├── audit.go                AuditEvent constants and AuditLog struct
│   │   ├── errors.go               Sentinel errors for all domain failures
│   │   └── repository.go           Repository interfaces owned by domain layer
│   ├── migrations/
│   │   ├── migrations.go           go:embed entry point for SQL files
│   │   └── sql/
│   │       ├── 001_users.sql       Users table schema
│   │       ├── 002_sessions.sql    Sessions table with token_hash indexes
│   │       └── 003_audit_logs.sql  Audit log table with event indexes
│   ├── repository/
│   │   └── sqlite/
│   │       ├── db.go               Connection, WAL mode, migration runner
│   │       ├── user_repo.go        UserRepository SQLite implementation
│   │       ├── session_repo.go     SessionRepository SQLite implementation
│   │       └── audit_repo.go       AuditRepository — fire and forget
│   ├── service/
│   │   ├── auth.go                 Register, Login, lockout, password strength
│   │   ├── session.go              Create, Validate, Invalidate sessions
│   │   └── totp.go                 Generate, Enable, Disable, Verify TOTP
│   ├── cli/
│   │   ├── app.go                  readline loop, state machine, routing
│   │   ├── pre_login.go            register and login command handlers
│   │   ├── post_login.go           whoami, 2FA, logout command handlers
│   │   └── display.go              All terminal output — tables, colors, formatting
│   └── logger/
│       └── logger.go               Dual slog handler — JSON file and text stdout
├── tests/
│   ├── auth_test.go                Password strength and hashing unit tests
│   └── totp_test.go                TOTP generation and verification unit tests
├── migrations/                     Original SQL files (also embedded via internal/migrations)
├── .env.example                    All environment variables documented with defaults
├── .dockerignore                   Excludes secrets and build artifacts from image
├── .gitignore                      Excludes .env, binaries, runtime data from git
├── Dockerfile                      Multi-stage distroless build
├── docker-compose.yml              Service definition with persistent named volumes
├── Makefile                        Developer workflow targets
├── LICENSE                         MIT License
└── README.md                       This document
```

---

## 🔧 Local Development

```bash
# Run locally without Docker
make run

# Build binary
make build
./bin/warden

# Run all tests with race detector
make test

# Generate HTML coverage report
make test-cover

# Run linter
make lint

# Clean build artifacts
make clean
```

### Available Make Targets

| Target | Description |
|---|---|
| `make build` | Compile binary to `bin/warden` |
| `make run` | Run locally with `go run` |
| `make test` | Run all tests with race detector |
| `make test-cover` | Generate HTML test coverage report |
| `make docker-up` | Build and start with docker-compose |
| `make docker-down` | Stop containers and remove volumes |
| `make lint` | Run `go vet` across all packages |
| `make clean` | Remove build artifacts |

---

## ✅ Requirements Coverage

| Requirement | Status | Implementation |
|---|---|---|
| User registration | ✅ | `register` command with username and password validation |
| Login with password | ✅ | `login` command with Argon2id verification |
| Optional TOTP 2FA | ✅ | `enable-2fa` and `disable-2fa` with Google Authenticator |
| Secure password storage | ✅ | Argon2id with random salt, OWASP recommended |
| Account lockout | ✅ | Progressive delays then 15-minute hard lockout |
| Session management | ✅ | Configurable timeout, DB-backed, validated per command |
| SQLite persistence | ✅ | WAL mode, named Docker volume, survives restarts |
| Interactive prompt | ✅ | readline with command history and tab completion |
| Clear error messages | ✅ | Colored pterm output with descriptive messages |
| Help command | ✅ | Context-aware — different commands before and after login |
| whoami command | ✅ | Full profile table auto-displayed on login |
| Docker container | ✅ | Multi-stage distroless build |
| README documentation | ✅ | This document |
| Database schema | ✅ | Three migration files, run automatically on startup |
| Unit tests | ✅ | 10 tests covering password strength and TOTP |

---

## 🔮 What I Would Add With More Time

- **WebAuthn / FIDO2** hardware security key support as a second authentication factor
- **IP-based rate limiting** to complement per-user account lockout for distributed attacks
- **Multi-device session management** with the ability to view and revoke individual sessions
- **TOTP backup codes** generated at 2FA setup for emergency account recovery
- **Admin CLI commands** to view audit logs, unlock accounts, and list active sessions
- **Automated integration tests** using an in-memory SQLite database for full flow coverage
- **Session refresh** to extend expiry on active use without requiring re-login

---

## 👨‍💻 Author

**Adesh Deshmukh**

[![GitHub](https://img.shields.io/badge/GitHub-AdeshDeshmukh-181717?style=flat&logo=github)](https://github.com/AdeshDeshmukh)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-adesh--deshmukh-0077B5?style=flat&logo=linkedin)](https://www.linkedin.com/in/adesh-deshmukh/)

---

## 📄 License

MIT License — Copyright © 2025 Adesh Deshmukh

See [LICENSE](LICENSE) for full terms.