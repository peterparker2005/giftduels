-- name: GetGiftByID :one
SELECT *
  FROM gifts
 WHERE id = $1;

-- name: GetUserGifts :many
SELECT *
  FROM gifts
 WHERE owner_telegram_id = $1
 ORDER BY created_at DESC
 LIMIT  $2
 OFFSET $3;

-- name: UpdateGiftStatus :one
UPDATE gifts 
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateGiftOwner :one
UPDATE gifts 
SET owner_telegram_id = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkGiftForWithdrawal :one
UPDATE gifts 
SET status = 'withdraw_pending', withdraw_requested = NOW(), updated_at = NOW()
WHERE id = $1 AND status = 'owned'
RETURNING *;

-- name: CompleteGiftWithdrawal :one
UPDATE gifts 
SET status = 'withdrawn', withdrawn_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status = 'withdraw_pending'
RETURNING *;

-- name: StakeGiftForGame :one
UPDATE gifts 
SET status = 'in_game', updated_at = NOW()
WHERE id = $1 AND status = 'owned'
RETURNING *;

-- name: CreateGiftEvent :one
INSERT INTO gift_events (
    gift_id,
    from_user_id,
    to_user_id
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetGiftEvents :many
SELECT * FROM gift_events
WHERE gift_id = $1
ORDER BY occurred_at DESC
LIMIT $2
OFFSET $3;

-- name: CreateGift :one
INSERT INTO gifts (
    id,
    telegram_gift_id,
    title,
    slug,
    owner_telegram_id,
    upgrade_message_id,
    ton_price,
    collectible_id,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: SaveGiftWithPrice :one
UPDATE gifts 
SET ton_price = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateGiftAttribute :one
INSERT INTO gift_attributes (
    gift_id,
    type,
    name,
    rarity
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;
