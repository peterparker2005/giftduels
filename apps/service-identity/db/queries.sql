-- name: GetUserByTelegramID :one
SELECT * FROM users WHERE telegram_id = $1 LIMIT 1;

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
