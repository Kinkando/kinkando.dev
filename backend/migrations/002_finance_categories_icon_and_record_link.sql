-- migrate:up
ALTER TABLE finance_categories ADD COLUMN icon TEXT NOT NULL DEFAULT '';
ALTER TABLE finance_records   ADD COLUMN category_id UUID REFERENCES finance_categories(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_finance_records_category ON finance_records (category_id);

-- migrate:down
DROP INDEX  IF EXISTS idx_finance_records_category;
ALTER TABLE finance_records   DROP COLUMN category_id;
ALTER TABLE finance_categories DROP COLUMN icon;
