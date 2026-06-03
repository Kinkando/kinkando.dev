-- migrate:up
ALTER TABLE health_profiles ADD COLUMN birthdate DATE;
ALTER TABLE health_profiles DROP COLUMN age;

-- migrate:down
ALTER TABLE health_profiles ADD COLUMN age INT;
ALTER TABLE health_profiles DROP COLUMN birthdate;
