-- migrate:up
ALTER TABLE workout_sessions ADD COLUMN completed_at TIMESTAMPTZ;

-- migrate:down
ALTER TABLE workout_sessions DROP COLUMN completed_at;
