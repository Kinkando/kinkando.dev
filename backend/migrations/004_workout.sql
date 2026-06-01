-- migrate:up

CREATE TABLE IF NOT EXISTS workout_presets (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_workout_presets_user ON workout_presets (user_id);

CREATE TABLE IF NOT EXISTS workout_preset_exercises (
    id               UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    preset_id        UUID    NOT NULL REFERENCES workout_presets(id) ON DELETE CASCADE,
    section          TEXT    NOT NULL DEFAULT 'main',
    order_index      INT     NOT NULL DEFAULT 0,
    name             TEXT    NOT NULL,
    target_muscles   TEXT,
    instructions     TEXT,
    sets             INT,
    reps             INT,
    duration_seconds INT,
    rest_seconds     INT,
    weight_kg        NUMERIC(6,2),
    equipment        TEXT,
    notes            TEXT
);
CREATE INDEX idx_workout_preset_exercises_preset ON workout_preset_exercises (preset_id, order_index);

CREATE TABLE IF NOT EXISTS workout_schedule (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week INT         NOT NULL,
    preset_id   UUID        NOT NULL REFERENCES workout_presets(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_workout_schedule_user_day UNIQUE (user_id, day_of_week)
);

CREATE TABLE IF NOT EXISTS workout_sessions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preset_id        UUID        REFERENCES workout_presets(id) ON DELETE SET NULL,
    name             TEXT        NOT NULL,
    type             TEXT        NOT NULL,
    performed_at     DATE        NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes INT,
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_workout_sessions_user_date ON workout_sessions (user_id, performed_at);

CREATE TABLE IF NOT EXISTS workout_session_exercises (
    id                      UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id              UUID    NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    section                 TEXT    NOT NULL DEFAULT 'main',
    order_index             INT     NOT NULL DEFAULT 0,
    name                    TEXT    NOT NULL,
    target_muscles          TEXT,
    instructions            TEXT,
    target_sets             INT,
    target_reps             INT,
    target_duration_seconds INT,
    rest_seconds            INT,
    actual_sets             INT,
    actual_reps             INT,
    actual_duration_seconds INT,
    weight_kg               NUMERIC(6,2),
    completed               BOOLEAN NOT NULL DEFAULT false,
    notes                   TEXT
);
CREATE INDEX idx_workout_session_exercises_session ON workout_session_exercises (session_id, order_index);

-- migrate:down
DROP TABLE IF EXISTS workout_session_exercises;
DROP TABLE IF EXISTS workout_sessions;
DROP TABLE IF EXISTS workout_schedule;
DROP TABLE IF EXISTS workout_preset_exercises;
DROP TABLE IF EXISTS workout_presets;
