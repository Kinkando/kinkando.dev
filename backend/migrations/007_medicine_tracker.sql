-- migrate:up
CREATE TABLE IF NOT EXISTS medicines (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                TEXT         NOT NULL,
    generic_name        TEXT,
    description         TEXT,
    stock_quantity      NUMERIC(10,2) NOT NULL DEFAULT 0,
    stock_unit          TEXT         NOT NULL,                          -- tablet|capsule|sachet|ml|...
    dosage_amount       NUMERIC(10,2) NOT NULL,
    dosage_unit         TEXT,
    frequency_type      TEXT         NOT NULL,                          -- daily|weekly|as_needed|custom
    frequency_value     INT,
    timing              TEXT,                                           -- before_meal|after_meal|before_breakfast|before_bed|anytime
    start_date          DATE,
    end_date            DATE,
    low_stock_threshold NUMERIC(10,2) NOT NULL DEFAULT 7,
    note                TEXT,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    archived_at         TIMESTAMPTZ
);
CREATE INDEX idx_medicines_user ON medicines (user_id, archived_at);

CREATE TABLE IF NOT EXISTS medicine_intakes (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    medicine_id    UUID          NOT NULL REFERENCES medicines(id) ON DELETE CASCADE,
    user_id        UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    medicine_name  TEXT          NOT NULL,                              -- snapshot at time of intake
    taken_at       TIMESTAMPTZ   NOT NULL DEFAULT now(),
    quantity_taken NUMERIC(10,2) NOT NULL,
    stock_before   NUMERIC(10,2) NOT NULL,
    stock_after    NUMERIC(10,2) NOT NULL,
    status         TEXT          NOT NULL DEFAULT 'taken',              -- taken|skipped|missed
    note           TEXT,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_medicine_intakes_user_date ON medicine_intakes (user_id, taken_at);
CREATE INDEX idx_medicine_intakes_medicine  ON medicine_intakes (medicine_id);

CREATE TABLE IF NOT EXISTS medicine_stock_adjustments (
    id           UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    medicine_id  UUID          NOT NULL REFERENCES medicines(id) ON DELETE CASCADE,
    user_id      UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type         TEXT          NOT NULL,                                -- add|remove|correction
    quantity     NUMERIC(10,2) NOT NULL,
    stock_before NUMERIC(10,2) NOT NULL,
    stock_after  NUMERIC(10,2) NOT NULL,
    reason       TEXT,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_medicine_stock_adj_user_date ON medicine_stock_adjustments (user_id, created_at);
CREATE INDEX idx_medicine_stock_adj_medicine  ON medicine_stock_adjustments (medicine_id);

-- migrate:down
DROP TABLE IF EXISTS medicine_stock_adjustments;
DROP TABLE IF EXISTS medicine_intakes;
DROP TABLE IF EXISTS medicines;
