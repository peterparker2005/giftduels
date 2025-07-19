-- Migration: telegram_message_id_rolls
-- Created at: 2025-07-19 14:46:34
-- Description: Add your migration description here

ALTER TABLE duel_rolls
  ADD COLUMN telegram_message_id INT NOT NULL DEFAULT 0;

ALTER TABLE duel_rolls
  ALTER COLUMN telegram_message_id DROP DEFAULT;
