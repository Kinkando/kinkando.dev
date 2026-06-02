-- migrate:up
CREATE TABLE IF NOT EXISTS health_food_logs (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT         NOT NULL,
    meal_type   TEXT         NOT NULL,  -- breakfast|lunch|dinner|snack
    calories    INT,
    protein_g   NUMERIC(6,2),
    carbs_g     NUMERIC(6,2),
    fat_g       NUMERIC(6,2),
    notes       TEXT,
    consumed_at DATE         NOT NULL DEFAULT CURRENT_DATE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_health_food_logs_user_date ON health_food_logs (user_id, consumed_at);

CREATE TABLE IF NOT EXISTS health_sleep_logs (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ  NOT NULL,  -- bedtime
    ended_at   TIMESTAMPTZ  NOT NULL,  -- wake time
    score      INT,                    -- 0–100 (Samsung Health / Galaxy Watch)
    notes      TEXT,
    logged_at  DATE         NOT NULL DEFAULT CURRENT_DATE,  -- night-of date for grouping
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_health_sleep_logs_user_date ON health_sleep_logs (user_id, logged_at);

-- migrate:down
DROP TABLE IF EXISTS health_sleep_logs;
DROP TABLE IF EXISTS health_food_logs;
