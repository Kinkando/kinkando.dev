-- migrate:up
CREATE TABLE IF NOT EXISTS notification_settings (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    line_enabled        BOOLEAN     NOT NULL DEFAULT false,
    discord_enabled     BOOLEAN     NOT NULL DEFAULT false,
    discord_webhook_url TEXT,
    web_push_enabled    BOOLEAN     NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fcm_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fcm_tokens_user ON fcm_tokens (user_id);

-- migrate:down
DROP INDEX IF EXISTS idx_fcm_tokens_user;
DROP TABLE IF EXISTS fcm_tokens;
DROP TABLE IF EXISTS notification_settings;
