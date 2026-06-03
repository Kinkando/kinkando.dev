-- migrate:up

-- Nullable per-user LINE user ID; unique among non-NULL values.
ALTER TABLE users ADD COLUMN line_id TEXT;
CREATE UNIQUE INDEX users_line_id_key ON users (line_id) WHERE line_id IS NOT NULL;

-- Pending one-time link codes (user sends "LINK <code>" to the bot).
CREATE TABLE line_link_codes (
    code       TEXT        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX line_link_codes_user_id_idx ON line_link_codes (user_id);

-- migrate:down
DROP TABLE IF EXISTS line_link_codes;
DROP INDEX IF EXISTS users_line_id_key;
ALTER TABLE users DROP COLUMN IF EXISTS line_id;
