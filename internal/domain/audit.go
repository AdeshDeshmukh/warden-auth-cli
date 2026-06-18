package domain

import "time"

type AuditEvent string

const (
	EventUserRegistered AuditEvent = "USER_REGISTERED"
	EventLoginSuccess   AuditEvent = "LOGIN_SUCCESS"
	EventLoginFailed    AuditEvent = "LOGIN_FAILED"
	EventAccountLocked  AuditEvent = "ACCOUNT_LOCKED"
	EventTOTPEnabled    AuditEvent = "TOTP_ENABLED"
	EventTOTPDisabled   AuditEvent = "TOTP_DISABLED"
	EventSessionCreated AuditEvent = "SESSION_CREATED"
	EventSessionExpired AuditEvent = "SESSION_EXPIRED"
	EventLogout         AuditEvent = "LOGOUT"
)

type AuditLog struct {
	ID        int64
	UserID    string
	Username  string
	Event     AuditEvent
	Detail    string
	CreatedAt time.Time
}