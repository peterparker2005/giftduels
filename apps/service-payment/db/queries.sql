-- name: CreateUserBalance :one
INSERT INTO user_balances (
    telegram_user_id,
    ton_balance
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
UPDATE user_balances SET ton_balance = ton_balance - $1 WHERE telegram_user_id = $2;

-- name: AddUserBalance :exec
UPDATE user_balances SET ton_balance = ton_balance + $1 WHERE telegram_user_id = $2;