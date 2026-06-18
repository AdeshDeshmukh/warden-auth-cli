# 🔐 Warden Auth CLI

> A production-grade, containerized command-line authentication system built in Go.
> Featuring Argon2id password hashing, TOTP two-factor authentication, structured audit
> logging, and a distroless Docker deployment.

---

## ✨ Highlights

- **Argon2id** password hashing — OWASP 2023 recommended, memory-hard algorithm
- **SHA256-hashed session tokens** — raw token never stored on disk or in database
- **Progressive account lockout** — escalating delays before hard lockout with exact countdown
- **TOTP two-factor authentication** — Google Authenticator and Authy compatible
- **Structured JSON audit logs** — every security event recorded, SIEM-ready format
- **Distroless Docker image** — zero shell, zero OS tools, minimal attack surface
- **Clean five-layer architecture** — domain, repository, service, CLI, config fully separated
- **Persistent storage** — SQLite with WAL mode, survives container restarts

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Warden Auth CLI                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   ┌─────────┐    ┌──────────┐    ┌────────────────┐    │
│   │   CLI   │───▶│ Service  │───▶│  Repository    │    │
│   │  Layer  │    │  Layer   │    │  Layer         │    │
│   └─────────┘    └──────────┘    └────────────────┘    │
│        │               │                  │             │
│        ▼               ▼                  ▼             │
│   ┌─────────┐    ┌──────────┐    ┌────────────────┐    │
│   │ Display │    │  Domain  │    │  SQLite + WAL  │    │
│   │  pterm  │    │  Models  │    │  (Persistent)  │    │
│   └─────────┘    └──────────┘    └────────────────┘    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Package | Responsibility |
|---|---|---|
| Domain | `internal/domain` | Models, interfaces, sentinel errors — zero external imports |
| Repository | `internal/repository/sqlite` | All database operations, maps SQL errors to domain errors |
| Service | `internal/service` | Business logic, security rules, no SQL and no CLI code |
| CLI | `internal/cli` | User input, command routing, display output |
| Config | `internal/config` | Environment variable loading with defaults |
| Logger | `internal/logger` | Dual-output structured logging (JSON file + stdout) |

---

## 🚀 Quick Start

**Prerequisites:** Docker and Docker Compose installed.

```bash
git clone https://github.com/AdeshDeshmukh/warden-auth-cli.git
cd warden-auth-cli
cp .env.example .env
docker-compose up --build
```

The application starts immediately. No additional setup required.

---

## 📖 Commands Reference

### Before Login

| Command | Description |
|---|---|
| `register` | Create a new user account |
| `login` | Authenticate with username and password |
| `help` | Show all available commands |
| `exit` | Quit the application |

### After Login

| Command | Description |
|---|---|
| `whoami` | Display your profile and active session details |
| `enable-2fa` | Set up TOTP two-factor authentication |
| `disable-2fa` | Remove two-factor authentication |
| `logout` | End your current session |
| `help` | Show all available commands |

---

## 🖥️ Usage Examples

### Register a New Account

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

┌─────────────────────────────────────────────┐
│ Field           │ Value                     │
├─────────────────────────────────────────────┤
│ Username        │ adesh                     │
│ Registered      │ 2025-06-17                │
│ Last Login      │ Just now                  │
│ 2FA Status      │ Disabled                  │
│ Session Expires │ in 30m 0s                 │
│ Account Status  │ Active                    │
└─────────────────────────────────────────────┘
```

### Enable Two-Factor Authentication

```
warden [adesh] ❯ enable-2fa
ℹ️  Scan the QR code below with Google Authenticator or Authy:
ℹ️  Manual entry key: JBSWY3DPEHPK3PXP
  Enter the 6-digit code from your authenticator app to confirm: 847291
✅ 2FA enabled successfully. Your account is now protected.
```

### Account Lockout

```
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

All configuration is managed through environment variables. Copy `.env.example` to `.env` and modify as needed.

| Variable | Default | Description |
|---|---|---|
| `DB_PATH` | `./data/warden.db` | SQLite database file path |
| `LOG_PATH` | `./logs/warden.log` | JSON audit log file path |
| `SESSION_TIMEOUT` | `30m` | Session duration before auto-expiry |
| `MAX_FAILED_ATTEMPTS` | `5` | Failed login attempts before lockout |
| `LOCKOUT_DURATION` | `15m` | Account lockout duration |
| `ARGON_MEMORY` | `65536` | Argon2id memory parameter in KB (64MB) |
| `ARGON_ITERATIONS` | `3` | Argon2id iteration count |
| `ARGON_PARALLELISM` | `2` | Argon2id parallel threads |

---

## 🔒 Security Design Decisions

### Argon2id over bcrypt

bcrypt is CPU-hard only. Argon2id is both CPU-hard and memory-hard. The memory requirement (64MB per attempt by default) makes GPU-based brute force attacks impractical because GPU cores share limited memory bandwidth. OWASP has recommended Argon2id as the preferred password hashing algorithm since 2023. The parameters used here (memory=64MB, iterations=3, parallelism=2) meet OWASP minimum requirements.

### SHA256 session token hashing

Storing raw session tokens in a database is equivalent to storing plaintext passwords. If the database file is compromised, an attacker would immediately have all active session tokens. By storing `SHA256(rawToken)` instead, the raw token exists only in the running process memory and is never written to disk. SHA256 is a one-way function — the stored hash cannot be reversed to recover the original token.

### Progressive lockout over hard cutoff

A hard lockout at N attempts is easy to detect during testing and probe around. Progressive delays (2 second artificial sleep from the third attempt onward) make brute force attempts painful before the hard lockout triggers. The exact countdown timer shown on lockout informs legitimate users without revealing internal timing information to attackers.

### Single session per user

When a new login succeeds, all existing sessions for that user are deleted before the new session is created. This prevents session fixation attacks where an attacker plants a known session identifier and waits for the victim to authenticate with it. It also prevents stale sessions from accumulating in the database.

### Distroless final Docker image

Standard base images (ubuntu, alpine) contain shells, package managers, and hundreds of OS-level binaries. Each represents a potential attack vector. The `gcr.io/distroless/static-debian12` image contains only the compiled binary and the minimal runtime dependencies. There is no `/bin/sh`, making shell injection structurally impossible. There is no package manager, eliminating supply chain attack surface. The image runs as a non-root user by default.

### Pure Go SQLite driver

Using `modernc.org/sqlite` instead of the CGo-based `mattn/go-sqlite3` means `CGO_ENABLED=0` during the Docker build. This produces a fully static binary with no shared library dependencies, which is what makes the distroless final image possible. A CGo build would require glibc to be present in the final image, forcing the use of a larger base image.

### Audit logs as a first-class feature

Every authentication event writes simultaneously to the `audit_logs` SQLite table and to a structured JSON log file. The JSON format is directly ingestible by SIEM tools such as Splunk, Elastic Stack, and Datadog. For a cybersecurity company, audit trails are not a feature — they are the foundation of security operations. The `username` field is denormalized intentionally so that the audit trail remains intact even if a user account is deleted.

### TOTP code verified before secret is saved

When enabling two-factor authentication, the user must successfully enter a valid six-digit code before the secret is written to the database. This proves the user has correctly scanned the QR code and their authenticator app is producing valid codes. Without this verification step, a user could enable 2FA with a broken configuration and then be permanently unable to log in.

### Timing attack prevention on login

When a username does not exist, the login handler still runs the full Argon2id computation on a dummy value before returning an error. Without this, an attacker could enumerate valid usernames by measuring response times — a non-existent user would return faster than an existing one. With the dummy computation, response times are identical regardless of whether the username exists.

---

## 🗄️ Database Schema

### users

| Column | Type | Description |
|---|---|---|
| `id` | TEXT | UUID primary key |
| `username` | TEXT | Unique username |
| `password_hash` | TEXT | Argon2id encoded hash |
| `totp_secret` | TEXT | TOTP secret, null if 2FA disabled |
| `totp_enabled` | INTEGER | 1 if 2FA enabled, 0 otherwise |
| `failed_attempts` | INTEGER | Failed login counter, persists across restarts |
| `locked_until` | DATETIME | Lockout expiry, null if not locked |
| `last_login_at` | DATETIME | Timestamp of last successful login |
| `created_at` | DATETIME | Account creation timestamp |

### sessions

| Column | Type | Description |
|---|---|---|
| `id` | TEXT | UUID primary key |
| `user_id` | TEXT | Foreign key to users.id |
| `token_hash` | TEXT | SHA256 of raw session token |
| `expires_at` | DATETIME | Session expiry timestamp |
| `created_at` | DATETIME | Session creation timestamp |

### audit_logs

| Column | Type | Description |
|---|---|---|
| `id` | INTEGER | Auto-increment primary key |
| `user_id` | TEXT | Foreign key to users.id, nullable |
| `username` | TEXT | Denormalized for log integrity |
| `event` | TEXT | Event type constant |
| `detail` | TEXT | JSON metadata string |
| `created_at` | DATETIME | Event timestamp |

---

## 📊 Audit Events Reference

| Event | Triggered When |
|---|---|
| `USER_REGISTERED` | New account created successfully |
| `LOGIN_SUCCESS` | Password and 2FA verified, session created |
| `LOGIN_FAILED` | Invalid password or account locked attempt |
| `ACCOUNT_LOCKED` | Failed attempts reached maximum threshold |
| `TOTP_ENABLED` | Two-factor authentication successfully enabled |
| `TOTP_DISABLED` | Two-factor authentication removed |
| `SESSION_CREATED` | New session token issued after login |
| `SESSION_EXPIRED` | Session validated but found to be past expiry |
| `LOGOUT` | User explicitly ended their session |

---

## 🐳 Docker Details

### Multi-Stage Build

**Stage 1 (builder):** Uses `golang:1.22-alpine` to compile a fully static binary with `CGO_ENABLED=0` and `-ldflags="-w -s"` to strip debug information and reduce binary size.

**Stage 2 (final):** Uses `gcr.io/distroless/static-debian12` containing only the compiled binary. Runs as `nonroot:nonroot` user. No shell, no package manager, no OS utilities.

### Persistent Volumes

| Volume | Mount | Contents |
|---|---|---|
| `db_data` | `/app/data` | SQLite database file |
| `log_data` | `/app/logs` | JSON audit log files |

Both volumes survive container restarts and rebuilds. The database is never lost when the container is stopped or updated.

---

## 📁 Project Structure

```
warden-auth-cli/
├── cmd/
│   └── main.go                        Entry point, dependency wiring, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go                  Environment variable loading with defaults
│   ├── domain/
│   │   ├── user.go                    User model with IsLocked and LockoutRemaining methods
│   │   ├── session.go                 Session model with IsExpired and TimeRemaining methods
│   │   ├── audit.go                   AuditEvent constants and AuditLog struct
│   │   ├── errors.go                  Sentinel errors for all domain failures
│   │   └── repository.go             Repository interfaces owned by domain layer
│   ├── migrations/
│   │   ├── migrations.go              go:embed directive for SQL files
│   │   └── sql/
│   │       ├── 001_users.sql          Users table with all security fields
│   │       ├── 002_sessions.sql       Sessions table with token_hash and indexes
│   │       └── 003_audit_logs.sql     Audit log table with event indexes
│   ├── repository/
│   │   └── sqlite/
│   │       ├── db.go                  Connection, WAL mode, migration runner
│   │       ├── user_repo.go           UserRepository implementation
│   │       ├── session_repo.go        SessionRepository implementation
│   │       └── audit_repo.go         AuditRepository implementation
│   ├── service/
│   │   ├── auth.go                    Register, Login, password hashing and lockout logic
│   │   ├── session.go                 Session creation, validation, and invalidation
│   │   └── totp.go                    TOTP generation, verification, enable and disable
│   ├── cli/
│   │   ├── app.go                     readline loop, state machine, command routing
│   │   ├── pre_login.go               register and login handlers
│   │   ├── post_login.go              whoami, 2FA, and logout handlers
│   │   └── display.go                 All terminal output, tables, colors, formatting
│   └── logger/
│       └── logger.go                  Dual slog handler (JSON file + text stdout)
├── tests/
│   ├── auth_test.go                   Password strength and hashing tests
│   └── totp_test.go                   TOTP generation and verification tests
├── .env.example                       All environment variables with defaults documented
├── .dockerignore                      Excludes sensitive and unnecessary files from image
├── .gitignore                         Excludes secrets, binaries, and runtime data
├── Dockerfile                         Multi-stage distroless build
├── docker-compose.yml                 App service with persistent volumes
├── Makefile                           Developer workflow targets
└── README.md                          This document
```

---

## 🧪 Running Tests

```bash
make test
```

```bash
make test-cover
```

The coverage report opens as an HTML file showing which lines are covered.

---

## 🔧 Local Development

```bash
cp .env.example .env
make run
```

Build binary:

```bash
make build
./bin/warden
```

Run linter:

```bash
make lint
```

---

## 🔮 What I Would Add With More Time

- **WebAuthn / FIDO2** hardware security key support as a second factor
- **IP-based rate limiting** to complement per-user account lockout
- **Multi-device session management** with the ability to view and revoke individual sessions
- **TOTP backup codes** for emergency account recovery
- **Admin CLI commands** to view audit logs and manually unlock accounts
- **Automated integration tests** using an in-memory SQLite database
- **Session refresh** to extend expiry on active use without requiring re-login

---

## 📄 License

MIT License — Copyright (c) 2025 Adesh Deshmukh

See [LICENSE](LICENSE) for full terms.