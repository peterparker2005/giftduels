-- Migration: transaction_metadata (DOWN)
-- Created at: 2025-07-10 23:39:29
-- Description: Rollback for transaction_metadata

-- Remove metadata column from user_transactions table
ALTER TABLE user_transactions
DROP COLUMN metadata;
