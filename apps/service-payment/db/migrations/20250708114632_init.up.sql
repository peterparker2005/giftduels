-- Migration: init
-- Created at: 2025-07-08 11:46:32
-- Description: Add your migration description here

CREATE TYPE transaction_reason AS ENUM (
	'withdraw', 'refund'
);

-- Create user_balances table
CREATE TABLE user_balances (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	telegram_user_id BIGINT NOT NULL,
	ton_amount FLOAT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create user_transactions table
CREATE TABLE user_transactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	telegram_user_id BIGINT NOT NULL,
	amount FLOAT NOT NULL,
	reason transaction_reason NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);