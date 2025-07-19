-- Migration: default_duel_status (DOWN)
-- Created at: 2025-07-19 14:46:04
-- Description: Rollback for default_duel_status

ALTER TABLE duels ALTER COLUMN status DROP DEFAULT;