-- name: CreateUserBalance :one
INSERT INTO user_balances (
    telegram_user_id,
    ton_amount
) VALUES (
    $1, $2
)
RETURNING *;

-- name: CreateUserTransaction :one
INSERT INTO user_transactions (
    telegram_user_id,
    amount,
    reason
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetUserBalance :one
SELECT * FROM user_balances WHERE telegram_user_id = $1;

-- name: SpendUserBalance :exec
UPDATE user_balances SET ton_amount = ton_amount - $1 WHERE telegram_user_id = $2;

-- name: AddUserBalance :exec
UPDATE user_balances SET ton_amount = ton_amount + $1 WHERE telegram_user_id = $2;

-- name: CreateDeposit :one
INSERT INTO deposits (telegram_user_id, amount_nano, payload, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDepositByPayload :one
SELECT * FROM deposits
WHERE payload = $1;

-- name: SetDepositTransaction :one
UPDATE deposits
SET
    status = 'received',
    tx_hash = $2,
    tx_lt = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetTonCursor :one
SELECT last_lt
FROM ton_cursors
WHERE network = $1
  AND wallet_address = $2;

-- name: UpsertTonCursor :exec
INSERT INTO ton_cursors (network, wallet_address, last_lt)
VALUES ($1, $2, $3)
ON CONFLICT (network, wallet_address) DO UPDATE
  SET last_lt    = EXCLUDED.last_lt,
      updated_at = now();
