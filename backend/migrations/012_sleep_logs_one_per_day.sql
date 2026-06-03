-- migrate:up
-- Remove duplicate (user_id, logged_at) rows before adding the unique constraint.
-- Keep the most recently created row for each date; delete the rest.
DELETE FROM health_sleep_logs
WHERE id NOT IN (
    SELECT DISTINCT ON (user_id, logged_at) id
    FROM health_sleep_logs
    ORDER BY user_id, logged_at, created_at DESC
);

ALTER TABLE health_sleep_logs
    ADD CONSTRAINT uq_health_sleep_logs_user_date UNIQUE (user_id, logged_at);

-- migrate:down
ALTER TABLE health_sleep_logs
    DROP CONSTRAINT IF EXISTS uq_health_sleep_logs_user_date;
