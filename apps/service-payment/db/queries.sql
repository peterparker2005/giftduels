-- name: CreateUserBalance :one
INSERT INTO user_balances (
    telegram_user_id,
    ton_amount
) VALUES (
    $1, $2
)
RETURNING *;

-- name: CreateTransaction :one
INSERT INTO user_transactions (
    telegram_user_id,
    amount,
    reason,
    metadata
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM user_transactions WHERE id = $1;

-- name: GetUserBalance :one
SELECT * FROM user_balances WHERE telegram_user_id = $1;

-- name: SpendUserBalance :one
UPDATE user_balances 
SET ton_amount = ton_amount - $2 
WHERE telegram_user_id = $1 AND ton_amount >= $2
RETURNING *;

-- name: AddUserBalance :one
UPDATE user_balances 
SET ton_amount = ton_amount + $2 
WHERE telegram_user_id = $1
RETURNING *;

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

-- name: GetUserTransactions :many
SELECT * FROM user_transactions
WHERE telegram_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserTransactionsCount :one
SELECT COUNT(*) FROM user_transactions
WHERE telegram_user_id = $1;