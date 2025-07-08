CREATE TYPE deposit_status AS ENUM ('pending', 'received', 'confirmed', 'expired');

CREATE TABLE deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_user_id BIGINT NOT NULL,
    status deposit_status NOT NULL DEFAULT 'pending',
    amount_nano BIGINT NOT NULL,
    payload TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    tx_hash TEXT,
    tx_lt BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX deposits_user_id_idx ON deposits(telegram_user_id); 