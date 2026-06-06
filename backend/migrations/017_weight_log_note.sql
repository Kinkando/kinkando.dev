-- migrate:up
ALTER TABLE health_weight_logs ADD COLUMN note TEXT;

-- migrate:down
ALTER TABLE health_weight_logs DROP COLUMN note;
