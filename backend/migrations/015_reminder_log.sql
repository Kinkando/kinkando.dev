-- migrate:up

-- Generic dedup log for scheduled batch reminders that are NOT medicine-specific
-- (quest reminders, weight-log nudges, etc.).
-- medicine_reminder_log (migration 014) is left unchanged — it carries a medicine FK.
--
-- UNIQUE (user_id, domain, reminder_key) makes every cron run idempotent:
-- INSERT … ON CONFLICT DO NOTHING; RowsAffected() > 0 ⟹ newly sent.
CREATE TABLE IF NOT EXISTS reminder_log (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain       TEXT        NOT NULL,   -- 'quest_daily' | 'quest_weekly' | 'weight'
    reminder_key TEXT        NOT NULL,   -- e.g. '2026-06-04' (daily/weight) or week-start '2026-06-02' (weekly)
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, domain, reminder_key)
);
CREATE INDEX IF NOT EXISTS idx_reminder_log_user ON reminder_log (user_id);

-- migrate:down
DROP INDEX  IF EXISTS idx_reminder_log_user;
DROP TABLE  IF EXISTS reminder_log;
