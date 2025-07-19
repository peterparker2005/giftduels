-- Migration: trigger_updated_at (DOWN)
-- Created at: 2025-07-19 14:39:08
-- Description: Rollback for trigger_updated_at

DROP TRIGGER IF EXISTS trg_gifts_set_updated_at ON gifts;
DROP FUNCTION IF EXISTS set_updated_at_timestamp;