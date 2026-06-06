-- migrate:up

-- Per-user unlock state for achievements/badges. The badge catalog itself lives
-- in code (internal/achievement); only the unlock + timestamp is persisted here.
-- One row per (user, badge code); evaluation is idempotent via the unique constraint.
CREATE TABLE user_achievements (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code        TEXT        NOT NULL,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_user_achievement UNIQUE (user_id, code)
);
CREATE INDEX idx_user_achievements_user ON user_achievements (user_id);

-- migrate:down
DROP TABLE IF EXISTS user_achievements;
