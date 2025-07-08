-- name: CreateUser :one
INSERT INTO users (
    telegram_id,
    username,
    first_name,
    last_name,
    photo_url,
    language_code,
    allows_write_to_pm,
    is_premium
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByTelegramID :one
SELECT * FROM users WHERE telegram_id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET
    username = $2,
    first_name = $3,
    last_name = $4,
    photo_url = $5,
    language_code = $6,
    allows_write_to_pm = $7,
    is_premium = $8
WHERE telegram_id = $1 RETURNING *;

-- name: UpsertUser :one
INSERT INTO users (
    telegram_id,
    username,
    first_name,
    last_name,
    photo_url,
    language_code,
    allows_write_to_pm,
    is_premium
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (telegram_id) DO UPDATE SET
    username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    photo_url = EXCLUDED.photo_url,
    language_code = EXCLUDED.language_code,
    allows_write_to_pm = EXCLUDED.allows_write_to_pm,
    is_premium = EXCLUDED.is_premium,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;
