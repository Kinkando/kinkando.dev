-- migrate:up

CREATE TABLE quest_definitions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         TEXT        NOT NULL,                   -- daily | weekly
    source_type  TEXT        NOT NULL DEFAULT 'manual',  -- manual | medicine | workout
    title        TEXT        NOT NULL,
    description  TEXT        NOT NULL DEFAULT '',
    xp_reward    INT         NOT NULL DEFAULT 0,
    target_count INT         NOT NULL DEFAULT 1,
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_quest_definitions_user_type   ON quest_definitions (user_id, type);
CREATE INDEX idx_quest_definitions_user_source ON quest_definitions (user_id, source_type);

CREATE TABLE quest_completions (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quest_id     UUID        NOT NULL REFERENCES quest_definitions(id) ON DELETE CASCADE,
    period_start DATE        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_quest_completions_quest_period ON quest_completions (quest_id, period_start);
CREATE INDEX idx_quest_completions_user_period  ON quest_completions (user_id, period_start);

CREATE TABLE user_xp_events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quest_id     UUID        REFERENCES quest_definitions(id) ON DELETE SET NULL,
    quest_title  TEXT        NOT NULL,
    source       TEXT        NOT NULL,
    period_start DATE        NOT NULL,
    xp           INT         NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_xp_quest_period UNIQUE (quest_id, period_start)
);
CREATE INDEX idx_user_xp_events_user_created ON user_xp_events (user_id, created_at);

-- migrate:down
DROP TABLE IF EXISTS user_xp_events;
DROP TABLE IF EXISTS quest_completions;
DROP TABLE IF EXISTS quest_definitions;
