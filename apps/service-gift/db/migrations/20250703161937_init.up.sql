-- 1. статусы
CREATE TYPE gift_status AS ENUM (
  'owned', 'in_game',
  'withdraw_pending', 'withdrawn'
);

CREATE TYPE gift_attribute_type AS ENUM (
  'model', 'symbol', 'backdrop'
);

-- 2. инстансы (immutable кроме owner/status/цен)
CREATE TABLE gifts (
  id               UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
  telegram_gift_id BIGINT    NOT NULL,
  collectible_id   INT       NOT NULL,
  owner_telegram_id         BIGINT    NOT NULL,
  upgrade_message_id INT      NOT NULL,
  title            TEXT      NOT NULL,
  slug             TEXT      NOT NULL,
  price            FLOAT     NOT NULL,
  emoji_id         BIGINT    NOT NULL,
  status           gift_status NOT NULL DEFAULT 'owned',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  withdrawn_at     TIMESTAMPTZ
);

-- 3. атрибуты конкретного инстанса
CREATE TABLE gift_attributes (
  gift_id UUID          REFERENCES gifts(id) ON DELETE CASCADE,
  type    gift_attribute_type          NOT NULL,
  name    TEXT          NOT NULL,
  rarity_per_mille  INT           NOT NULL,
  PRIMARY KEY (gift_id, type)
);

-- 4. события (аудит)
CREATE TABLE gift_events (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  gift_id      UUID REFERENCES gifts(id) ON DELETE CASCADE,
  from_user_id BIGINT NULL,
  to_user_id   BIGINT NULL,
  action       TEXT  NOT NULL,
  source       TEXT  NULL, -- duel, jackpot, system
  related_game_id TEXT NULL,
  description  TEXT NULL,
  payload      JSONB NULL,
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE UNIQUE INDEX ux_active_tg_id
  ON gifts(telegram_gift_id)
  WHERE status <> 'withdrawn';
