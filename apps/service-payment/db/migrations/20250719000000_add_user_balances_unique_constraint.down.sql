-- Migration: add_user_balances_unique_constraint (DOWN)
-- Created at: 2025-07-19 00:00:00
-- Description: Rollback for add_user_balances_unique_constraint

-- Remove unique constraint on telegram_user_id
ALTER TABLE user_balances
DROP CONSTRAINT IF EXISTS user_balances_telegram_user_id_unique; 