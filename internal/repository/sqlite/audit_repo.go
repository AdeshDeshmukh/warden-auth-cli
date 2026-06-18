package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AdeshDeshmukh/warden-auth-cli/internal/domain"
)

type AuditRepo struct {
	db *sql.DB
}

func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Log(ctx context.Context, entry domain.AuditLog) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO audit_logs (user_id, username, event, detail, created_at)
		VALUES (?, ?, ?, ?, ?)`

	userID := sql.NullString{
		String: entry.UserID,
		Valid:  entry.UserID != "",
	}

	_, err := r.db.ExecContext(ctx, query,
		userID,
		entry.Username,
		string(entry.Event),
		entry.Detail,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}