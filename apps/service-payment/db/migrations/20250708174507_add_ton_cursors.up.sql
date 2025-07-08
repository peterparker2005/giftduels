-- Migration: add_ton_cursors
-- Created at: 2025-07-08 17:45:07
-- Description: Add your migration description here

CREATE TYPE ton_network AS ENUM ('mainnet', 'testnet');

CREATE TABLE ton_cursors (
  network        ton_network  NOT NULL DEFAULT 'mainnet',
  wallet_address TEXT         NOT NULL,
  last_lt        BIGINT       NOT NULL,
  updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
  PRIMARY KEY (network, wallet_address)
);