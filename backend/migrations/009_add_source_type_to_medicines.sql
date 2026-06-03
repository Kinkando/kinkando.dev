-- migrate:up
ALTER TABLE medicines ADD COLUMN source_type TEXT NOT NULL DEFAULT 'medication';  -- medication|supplement

-- migrate:down
ALTER TABLE medicines DROP COLUMN source_type;
