-- name: GetGiftByID :one
SELECT
  g.*,
  gc.id AS collection_id, gc.name AS collection_name, gc.short_name AS collection_short_name,
  gm.id AS model_id, gm.name AS model_name, gm.short_name AS model_short_name, gm.rarity_per_mille AS model_rarity,
  gb.id AS backdrop_id, gb.name AS backdrop_name, gb.short_name AS backdrop_short_name, gb.rarity_per_mille AS backdrop_rarity,
  gb.center_color, gb.edge_color, gb.pattern_color, gb.text_color,
  gs.id AS symbol_id, gs.name AS symbol_name, gs.short_name AS symbol_short_name, gs.rarity_per_mille AS symbol_rarity
FROM gifts g
JOIN gift_collections gc ON gc.id = g.collection_id
JOIN gift_models gm ON gm.id = g.model_id
JOIN gift_backdrops gb ON gb.id = g.backdrop_id
JOIN gift_symbols gs ON gs.id = g.symbol_id
WHERE g.id = $1;

-- name: GetGiftsByIDs :many
SELECT
  g.*,
  gc.id AS collection_id, gc.name AS collection_name, gc.short_name AS collection_short_name,
  gm.id AS model_id, gm.name AS model_name, gm.short_name AS model_short_name, gm.rarity_per_mille AS model_rarity,
  gb.id AS backdrop_id, gb.name AS backdrop_name, gb.short_name AS backdrop_short_name, gb.rarity_per_mille AS backdrop_rarity,
  gb.center_color, gb.edge_color, gb.pattern_color, gb.text_color,
  gs.id AS symbol_id, gs.name AS symbol_name, gs.short_name AS symbol_short_name, gs.rarity_per_mille AS symbol_rarity
FROM gifts g
JOIN gift_collections gc ON gc.id = g.collection_id
JOIN gift_models gm ON gm.id = g.model_id
JOIN gift_backdrops gb ON gb.id = g.backdrop_id
JOIN gift_symbols gs ON gs.id = g.symbol_id
WHERE g.id = ANY($1::uuid[]);

-- name: GetUserGifts :many
SELECT
  g.*,
  gc.id AS collection_id, gc.name AS collection_name, gc.short_name AS collection_short_name,
  gm.id AS model_id, gm.name AS model_name, gm.short_name AS model_short_name,
  gb.id AS backdrop_id, gb.name AS backdrop_name, gb.short_name AS backdrop_short_name,
  gs.id AS symbol_id, gs.name AS symbol_name, gs.short_name AS symbol_short_name
FROM gifts g
JOIN gift_collections gc ON gc.id = g.collection_id
JOIN gift_models gm ON gm.id = g.model_id
JOIN gift_backdrops gb ON gb.id = g.backdrop_id
JOIN gift_symbols gs ON gs.id = g.symbol_id
WHERE g.owner_telegram_id = $1
ORDER BY g.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserGiftsCount :one
SELECT COUNT(*)
FROM gifts
WHERE owner_telegram_id = $1;

-- name: GetUserActiveGifts :many
SELECT
  -- поля самого подарка
  g.id,
  g.telegram_gift_id,
  g.collectible_id,
  g.owner_telegram_id,
  g.upgrade_message_id,
  g.title,
  g.slug,
  g.price,
  g.status,
  g.created_at,
  g.updated_at,
  g.withdrawn_at,

  -- collection
  gc.id   AS collection_id,
  gc.name AS collection_name,
  gc.short_name AS collection_short_name,

  -- model (вместе с rarity)
  gm.id                 AS model_id,
  gm.name               AS model_name,
  gm.short_name         AS model_short_name,
  gm.rarity_per_mille   AS model_rarity,

  -- backdrop (с rarity и цветами)
  gb.id                 AS backdrop_id,
  gb.name               AS backdrop_name,
  gb.short_name         AS backdrop_short_name,
  gb.rarity_per_mille   AS backdrop_rarity,
  gb.center_color       AS backdrop_center_color,
  gb.edge_color         AS backdrop_edge_color,
  gb.pattern_color      AS backdrop_pattern_color,
  gb.text_color         AS backdrop_text_color,

  -- symbol (с rarity)
  gs.id               AS symbol_id,
  gs.name             AS symbol_name,
  gs.short_name       AS symbol_short_name,
  gs.rarity_per_mille AS symbol_rarity

FROM gifts g
  JOIN gift_collections gc ON gc.id = g.collection_id
  JOIN gift_models     gm ON gm.id = g.model_id
  JOIN gift_backdrops  gb ON gb.id = g.backdrop_id
  JOIN gift_symbols    gs ON gs.id = g.symbol_id
WHERE g.owner_telegram_id = $1
  AND g.status NOT IN ('withdrawn','withdraw_pending')
ORDER BY g.created_at DESC
LIMIT  $2
OFFSET $3;

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
