-- migrate:up
CREATE TABLE IF NOT EXISTS health_profiles (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    height     NUMERIC(5,2),
    age        INT,
    gender     TEXT,
    goal       TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS health_weight_logs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weight     NUMERIC(5,2) NOT NULL,
    logged_at  DATE        NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_health_weight_logs_user_date UNIQUE (user_id, logged_at)
);
CREATE INDEX idx_health_weight_logs_user_date ON health_weight_logs (user_id, logged_at);

-- migrate:down
DROP TABLE IF EXISTS health_weight_logs;
DROP TABLE IF EXISTS health_profiles;
