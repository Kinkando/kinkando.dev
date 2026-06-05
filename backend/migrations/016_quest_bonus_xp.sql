-- migrate:up

-- Deduplicate bonus XP rows (quest_id IS NULL) per user/source/period.
-- Regular per-quest rows are deduped by uq_xp_quest_period (quest_id, period_start);
-- that constraint treats NULLs as distinct, so bonus rows need their own guard.
CREATE UNIQUE INDEX uq_xp_bonus_period
    ON user_xp_events (user_id, source, period_start)
    WHERE quest_id IS NULL;

-- migrate:down
DROP INDEX IF EXISTS uq_xp_bonus_period;
