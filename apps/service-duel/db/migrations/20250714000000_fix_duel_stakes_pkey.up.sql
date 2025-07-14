-- Migration: fix_duel_stakes_pkey
-- Created at: 2025-07-14 22:54:00
-- Description: Fix primary key constraint for duel_stakes to allow multiple stakes per user

-- Drop the existing primary key constraint
ALTER TABLE duel_stakes DROP CONSTRAINT duel_stakes_pkey;

-- Add the new primary key constraint that includes gift_id
ALTER TABLE duel_stakes ADD PRIMARY KEY (duel_id, telegram_user_id, gift_id); 