-- Migration: default_duel_status
-- Created at: 2025-07-19 14:46:04
-- Description: Add your migration description here

ALTER TABLE duels ALTER COLUMN status SET DEFAULT 'waiting_for_opponent';
