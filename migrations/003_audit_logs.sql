CREATE TABLE IF NOT EXISTS audit_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     TEXT REFERENCES users(id) ON DELETE SET NULL,
    username    TEXT NOT NULL,
    event       TEXT NOT NULL,
    detail      TEXT DEFAULT '{}',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_user
    ON audit_logs(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_event
    ON audit_logs(event);

CREATE INDEX IF NOT EXISTS idx_audit_time
    ON audit_logs(created_at DESC);