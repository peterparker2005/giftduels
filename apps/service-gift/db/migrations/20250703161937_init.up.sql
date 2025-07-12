-- Migration: init
-- Created at: ?
-- Description: Init schema for gifts with collections, models, backdrops, symbols, events

-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Enum type for gift status
CREATE TYPE gift_status AS ENUM (
  'owned',
  'in_game',
  'withdraw_pending',
  'withdrawn'
);

-- 2. Reference tables (must be created before gifts)

-- Collections
CREATE TABLE gift_collections (
  id          SERIAL      PRIMARY KEY,
  name        TEXT        NOT NULL UNIQUE,
  short_name  TEXT        NOT NULL UNIQUE
);

-- Models (belong to a specific collection)
CREATE TABLE gift_models (
  id               SERIAL      PRIMARY KEY,
  collection_id    INT         NOT NULL REFERENCES gift_collections(id) ON DELETE CASCADE,
  name             TEXT        NOT NULL,
  short_name       TEXT        NOT NULL,
  rarity_per_mille INT         NOT NULL,
  UNIQUE(collection_id, name),
  UNIQUE(collection_id, short_name)
);

-- Backdrops (can be used by any collection)
CREATE TABLE gift_backdrops (
  id               SERIAL      PRIMARY KEY,
  name             TEXT        NOT NULL UNIQUE,
  short_name       TEXT        NOT NULL UNIQUE,
  rarity_per_mille INT         NOT NULL,
  center_color     TEXT,
  edge_color       TEXT,
  pattern_color    TEXT,
  text_color       TEXT
);

-- Symbols (patterns, can be used by any collection)
CREATE TABLE gift_symbols (
  id               SERIAL      PRIMARY KEY,
  name             TEXT        NOT NULL UNIQUE,
  short_name       TEXT        NOT NULL UNIQUE,
  rarity_per_mille INT         NOT NULL
);

-- 3. Main gifts table
CREATE TABLE gifts (
  id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  telegram_gift_id    BIGINT        NOT NULL,
  collectible_id      INT           NOT NULL,
  owner_telegram_id   BIGINT        NOT NULL,
  upgrade_message_id  INT           NOT NULL,
  title               TEXT          NOT NULL,
  slug                TEXT          NOT NULL,
  price               FLOAT         NOT NULL,
  collection_id       INT           NOT NULL REFERENCES gift_collections(id),
  model_id            INT           NOT NULL REFERENCES gift_models(id),
  backdrop_id         INT           NOT NULL REFERENCES gift_backdrops(id),
  symbol_id           INT           NOT NULL REFERENCES gift_symbols(id),
  status              gift_status   NOT NULL DEFAULT 'owned',
  created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  withdrawn_at        TIMESTAMPTZ
);

-- 4. Events (audit log)
CREATE TABLE gift_events (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  gift_id          UUID        NOT NULL REFERENCES gifts(id) ON DELETE CASCADE,
  from_user_id     BIGINT,
  to_user_id       BIGINT,
  action           TEXT        NOT NULL,
  game_mode        TEXT,
  related_game_id  TEXT,
  description      TEXT,
  payload          JSONB,
  occurred_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Unique index to enforce one active mapping per telegram_gift_id
CREATE UNIQUE INDEX ux_active_tg_id
  ON gifts(telegram_gift_id)
  WHERE status <> 'withdrawn';
