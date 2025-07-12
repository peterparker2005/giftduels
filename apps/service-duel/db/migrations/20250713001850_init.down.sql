-- Migration: init (DOWN)
-- Created at: 2025-07-13 00:18:50
-- Description: Rollback for init

-- delete trigger
DROP TRIGGER IF EXISTS duel_updated_at ON duels;

-- delete function
DROP FUNCTION IF EXISTS trg_update_timestamp();

-- delete tables (in order of dependency: first child → parent)
DROP TABLE IF EXISTS duel_stakes;
DROP TABLE IF EXISTS duel_rolls;
DROP TABLE IF EXISTS duel_rounds;
DROP TABLE IF EXISTS duel_participants;
DROP TABLE IF EXISTS duels;

-- delete index (if it didn't delete together with the table)
DROP INDEX IF EXISTS idx_duels_next_roll_deadline;

-- delete enum type
DROP TYPE IF EXISTS duel_status;

-- pgcrypto extension can be left, it can be used in other places
-- DROP EXTENSION IF EXISTS pgcrypto;
