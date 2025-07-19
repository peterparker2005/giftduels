-- Migration: telegram_message_id_rolls (DOWN)
-- Created at: 2025-07-19 14:46:34
-- Description: Rollback for telegram_message_id_rolls

ALTER TABLE duel_rolls DROP COLUMN telegram_message_id;