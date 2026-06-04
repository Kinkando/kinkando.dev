-- migrate:up

-- Reminder columns on medicines: when enabled, the cron batch job will send
-- dose-due and missed-dose notifications at the stored clock times (Asia/Bangkok).
-- reminder_times is a JSON array of "HH:MM" strings, e.g. '["08:00","20:00"]'.
ALTER TABLE medicines
    ADD COLUMN IF NOT EXISTS reminder_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS reminder_times   TEXT    NOT NULL DEFAULT '[]';

-- Dedup log — one row per (medicine_id, reminder_type, reminder_key) prevents
-- the cron worker from re-sending the same alert on repeated runs within the
-- same dedup window (daily for supply digests, per-slot for dose/missed).
CREATE TABLE IF NOT EXISTS medicine_reminder_log (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    medicine_id   UUID        NOT NULL REFERENCES medicines(id) ON DELETE CASCADE,
    reminder_type TEXT        NOT NULL,   -- low_stock | refill | dose | missed
    reminder_key  TEXT        NOT NULL,   -- e.g. '2026-06-04' or '2026-06-04#08:00' or 'missed#2026-06-04#08:00'
    sent_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (medicine_id, reminder_type, reminder_key)
);
CREATE INDEX IF NOT EXISTS idx_medicine_reminder_log_user ON medicine_reminder_log (user_id);

-- migrate:down
DROP INDEX  IF EXISTS idx_medicine_reminder_log_user;
DROP TABLE  IF EXISTS medicine_reminder_log;
ALTER TABLE medicines DROP COLUMN IF EXISTS reminder_enabled;
ALTER TABLE medicines DROP COLUMN IF EXISTS reminder_times;
