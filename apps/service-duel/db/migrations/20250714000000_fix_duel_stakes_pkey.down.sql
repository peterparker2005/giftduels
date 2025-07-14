-- Migration: fix_duel_stakes_pkey (down)
-- Created at: 2025-07-14 22:54:00
-- Description: Revert primary key constraint for duel_stakes

-- Drop the new primary key constraint
ALTER TABLE duel_stakes DROP CONSTRAINT duel_stakes_pkey;

-- Restore the original primary key constraint
ALTER TABLE duel_stakes ADD PRIMARY KEY (duel_id, telegram_user_id); 