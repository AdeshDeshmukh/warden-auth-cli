package domain

import "time"

type User struct {
	ID             string
	Username       string
	PasswordHash   string
	TOTPSecret     string
	TOTPEnabled    bool
	FailedAttempts int
	LockedUntil    *time.Time
	LastLoginAt    *time.Time
	CreatedAt      time.Time
}

func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}

func (u *User) LockoutRemaining() time.Duration {
	if !u.IsLocked() {
		return 0
	}
	return time.Until(*u.LockedUntil)
}