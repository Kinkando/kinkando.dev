// Package reminderlog provides a domain-agnostic dedup log for scheduled batch
// reminders. Each (user_id, domain, reminder_key) triple may only be inserted
// once — the unique constraint on the reminder_log table prevents re-sending.
package reminderlog

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/gen/kinkando/public/table"
)

// Domain identifies which reminder batch job owns a reminder_log entry.
type Domain string

const (
	DomainQuestDaily  Domain = "quest_daily"
	DomainQuestWeekly Domain = "quest_weekly"
	DomainWeight      Domain = "weight"
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
// key examples: "2026-06-04" (daily/weight), "2026-06-02" (weekly period start)
func (r *Repository) Log(ctx context.Context, userID uuid.UUID, domain Domain, key string) (bool, error) {
	stmt := table.ReminderLog.INSERT(
		table.ReminderLog.UserID,
		table.ReminderLog.Domain,
		table.ReminderLog.ReminderKey,
	).VALUES(
		postgres.UUID(userID),
		string(domain),
		key,
	).ON_CONFLICT(
		table.ReminderLog.UserID,
		table.ReminderLog.Domain,
		table.ReminderLog.ReminderKey,
	).DO_NOTHING()

	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return false, fmt.Errorf("reminder log: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
