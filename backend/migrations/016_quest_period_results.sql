-- migrate:up

-- Snapshot of every active quest's final state at the end of each daily/weekly period.
-- One row per quest per period; written by the quest-period-snapshot cron job (POST /api/v1/cron/quest-period-snapshot).
-- quest_id is nullable so the record survives if the quest definition is later deleted (mirrors user_xp_events).
-- The unique constraint on (quest_id, period_start) makes every cron run fully idempotent.
CREATE TABLE IF NOT EXISTS quest_period_results (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quest_id        UUID        REFERENCES quest_definitions(id) ON DELETE SET NULL,
    quest_title     TEXT        NOT NULL,                   -- snapshot of title at period end
    type            TEXT        NOT NULL,                   -- 'daily' | 'weekly'
    period_start    DATE        NOT NULL,
    target_count    INT         NOT NULL,
    completed_count INT         NOT NULL,
    completed       BOOLEAN     NOT NULL,                   -- completed_count >= target_count
    xp_reward       INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_quest_period_result UNIQUE (quest_id, period_start)
);
CREATE INDEX IF NOT EXISTS idx_quest_period_results_user_period
    ON quest_period_results (user_id, period_start);

-- migrate:down
DROP INDEX IF EXISTS idx_quest_period_results_user_period;
DROP TABLE IF EXISTS quest_period_results;
