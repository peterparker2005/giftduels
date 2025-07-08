-- Migration: init (DOWN)
-- Created at: 2025-07-08 11:46:32
-- Description: Rollback for init

DROP TABLE IF EXISTS user_balances;
DROP TABLE IF EXISTS user_transactions;

DROP TYPE IF EXISTS transaction_reason;