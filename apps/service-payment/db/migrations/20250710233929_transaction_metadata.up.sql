-- Migration: transaction_metadata
-- Created at: 2025-07-10 23:39:29
-- Description: Add your migration description here

-- Add metadata column to user_transactions table
ALTER TABLE user_transactions
ADD COLUMN metadata JSONB NULL;