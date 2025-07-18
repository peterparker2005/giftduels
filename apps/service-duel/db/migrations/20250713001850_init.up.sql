-- Migration: init
-- Created at: 2025-07-13 00:18:50
-- Description: Add your migration description here

-- Extension for UUID generation (Postgres 13+ uses pgcrypto)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enum type for duel status
CREATE TYPE duel_status AS ENUM (
  'waiting_for_opponent',
  'in_progress',
  'completed',
  'cancelled'
);

-- Main duels table
CREATE TABLE duels (
  id                   UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
  display_number       BIGINT           NOT NULL GENERATED ALWAYS AS IDENTITY,
  is_private           BOOLEAN          NOT NULL,
  max_players          INTEGER          NOT NULL CHECK (max_players BETWEEN 2 AND 4),
  max_gifts            INTEGER          NOT NULL CHECK (max_gifts BETWEEN 1 AND 10),
  winner_telegram_user_id BIGINT        NULL,
  next_roll_deadline   TIMESTAMPTZ      NULL,
  status               duel_status      NULL,
  created_at           TIMESTAMPTZ      NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ      NOT NULL DEFAULT now(),
  completed_at         TIMESTAMPTZ      NULL
);

-- Participants in each duel
CREATE TABLE duel_participants (
  duel_id              UUID             NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
  telegram_user_id     BIGINT           NOT NULL,
  is_creator           BOOLEAN          NOT NULL,
  PRIMARY KEY (duel_id, telegram_user_id)
);

-- Rounds within each duel
CREATE TABLE duel_rounds (
  duel_id              UUID             NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
  round_number         INTEGER          NOT NULL CHECK (round_number > 0),
  PRIMARY KEY (duel_id, round_number)
);

-- Individual rolls in each round
CREATE TABLE duel_rolls (
  duel_id                 UUID             NOT NULL,
  round_number            INTEGER          NOT NULL,
  telegram_user_id        BIGINT           NOT NULL,
  dice_value              SMALLINT         NOT NULL CHECK (dice_value BETWEEN 1 AND 6),
  rolled_at               TIMESTAMPTZ      NOT NULL,
  is_auto_rolled          BOOLEAN          NOT NULL DEFAULT FALSE,
  PRIMARY KEY (duel_id, round_number, telegram_user_id),
  FOREIGN KEY (duel_id, round_number)
    REFERENCES duel_rounds(duel_id, round_number)
      ON DELETE CASCADE
);

-- Stakes placed by participants
CREATE TABLE duel_stakes (
  duel_id              UUID             NOT NULL REFERENCES duels(id) ON DELETE CASCADE,
  telegram_user_id     BIGINT           NOT NULL,
  gift_id              UUID             NOT NULL,
  PRIMARY KEY (duel_id, telegram_user_id, gift_id)
);

-- Index to quickly find duels needing auto-roll checks
CREATE INDEX idx_duels_next_roll_deadline
  ON duels(next_roll_deadline)
  WHERE status = 'in_progress';

-- Trigger to update updated_at on row change
CREATE OR REPLACE FUNCTION trg_update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER duel_updated_at
  BEFORE UPDATE ON duels
  FOR EACH ROW
  EXECUTE PROCEDURE trg_update_timestamp();
