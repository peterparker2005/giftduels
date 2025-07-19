-- Migration: add_user_balances_unique_constraint
-- Created at: 2025-07-19 00:00:00
-- Description: Add unique constraint on telegram_user_id in user_balances table

-- Add unique constraint on telegram_user_id
ALTER TABLE user_balances
ADD CONSTRAINT user_balances_telegram_user_id_unique UNIQUE (telegram_user_id); 