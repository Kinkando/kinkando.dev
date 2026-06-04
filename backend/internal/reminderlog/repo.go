// Package reminderlog provides a domain-agnostic dedup log for scheduled batch
// reminders. Each (user_id, domain, reminder_key) triple may only be inserted
// once — the unique constraint on the reminder_log table prevents re-sending.
package reminderlog

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Repository wraps the reminder_log table with a single idempotent Log method.
type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Log inserts a reminder-log entry for (userID, domain, key).
// It returns true when the row was newly inserted (reminder not yet sent),
// or false when the ON CONFLICT suppressed the insert (already sent this period).
//
// domain examples: "quest_daily", "quest_weekly", "weight"
// key examples:    "2026-06-04" (daily/weight), "2026-06-02" (weekly period start)
func (r *Repository) Log(ctx context.Context, userID uuid.UUID, domain, key string) (bool, error) {
	const q = `
		INSERT INTO reminder_log (user_id, domain, reminder_key)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, domain, reminder_key) DO NOTHING`

	res, err := r.db.ExecContext(ctx, q, userID, domain, key)
	if err != nil {
		return false, fmt.Errorf("reminder log: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
