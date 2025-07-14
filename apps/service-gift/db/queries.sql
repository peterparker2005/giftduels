-- name: GetGiftByID :one
SELECT *
FROM gifts
WHERE id = $1;

-- name: GetGiftsByIDs :many
SELECT *
FROM gifts
WHERE id = ANY($1::uuid[]);

-- name: GetUserGifts :many
SELECT *
FROM gifts
WHERE owner_telegram_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserGiftsCount :one
SELECT COUNT(*)
FROM gifts
WHERE owner_telegram_id = $1;

-- name: GetUserActiveGifts :many
SELECT *
FROM gifts
WHERE owner_telegram_id = $1
  AND status NOT IN ('withdrawn', 'withdraw_pending')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserActiveGiftsCount :one
SELECT COUNT(*)
FROM gifts
WHERE owner_telegram_id = $1
  AND status NOT IN ('withdrawn', 'withdraw_pending');

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
SET status = 'withdraw_pending', updated_at = NOW()
WHERE id = $1 AND status = 'owned'
RETURNING *;

-- name: CancelGiftWithdrawal :one
UPDATE gifts 
SET status = 'owned', updated_at = NOW()
WHERE id = $1 AND status = 'withdraw_pending'
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
    to_user_id,
    related_game_id,
    game_mode,
    description,
    payload
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
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
    price,
    collectible_id,
    collection_id,
    model_id,
    backdrop_id,
    symbol_id,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: SaveGiftWithPrice :one
UPDATE gifts 
SET price = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetGiftModel :one
SELECT * FROM gift_models
WHERE id = $1;

-- name: GetGiftBackdrop :one
SELECT * FROM gift_backdrops
WHERE id = $1;

-- name: GetGiftSymbol :one
SELECT * FROM gift_symbols
WHERE id = $1;

-- name: GetGiftCollection :one
SELECT * FROM gift_collections
WHERE id = $1;

-- name: CreateCollection :one
INSERT INTO gift_collections (name, short_name)
VALUES ($1, $2)
RETURNING *;

-- name: CreateModel :one
INSERT INTO gift_models (collection_id, name, short_name, rarity_per_mille)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateBackdrop :one
INSERT INTO gift_backdrops (name, short_name, rarity_per_mille, center_color, edge_color, pattern_color, text_color)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CreateSymbol :one
INSERT INTO gift_symbols (name, short_name, rarity_per_mille)
VALUES ($1, $2, $3)
RETURNING *;

-- name: FindCollectionByName :one
SELECT * FROM gift_collections
WHERE name = $1;

-- name: FindModelByName :one
SELECT * FROM gift_models
WHERE name = $1;

-- name: FindBackdropByName :one
SELECT * FROM gift_backdrops
WHERE name = $1;

-- name: FindSymbolByName :one
SELECT * FROM gift_symbols
WHERE name = $1;

-- name: GetGiftCollectionsByIDs :many
SELECT *
FROM gift_collections
WHERE id = ANY($1::int[]);

-- name: GetGiftModelsByIDs :many
SELECT *
FROM gift_models
WHERE id = ANY($1::int[]);

-- name: GetGiftBackdropsByIDs :many
SELECT *
FROM gift_backdrops
WHERE id = ANY($1::int[]);

-- name: GetGiftSymbolsByIDs :many
SELECT *
FROM gift_symbols
WHERE id = ANY($1::int[]);
