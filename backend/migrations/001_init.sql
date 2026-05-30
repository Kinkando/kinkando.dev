-- migrate:up

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users (mirrors Firebase accounts; finance records FK into this table)
CREATE TABLE IF NOT EXISTS users (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid TEXT        NOT NULL UNIQUE,
    email        TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Finance records
CREATE TABLE IF NOT EXISTS finance_records (
    id         UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT           NOT NULL CHECK (type IN ('income', 'expense')),
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    category   TEXT           NOT NULL DEFAULT '',
    note       TEXT           NOT NULL DEFAULT '',
    date       DATE           NOT NULL,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_finance_records_user_date
    ON finance_records (user_id, date DESC);

-- Finance categories (user-defined palette)
CREATE TABLE IF NOT EXISTS finance_categories (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    type       TEXT        NOT NULL CHECK (type IN ('income', 'expense')),
    color      TEXT        NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name, type)
);

-- migrate:down

DROP INDEX  IF EXISTS idx_finance_records_user_date;
DROP TABLE  IF EXISTS finance_categories;
DROP TABLE  IF EXISTS finance_records;
DROP TABLE  IF EXISTS users;
