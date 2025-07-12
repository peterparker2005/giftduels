-- name: GetDuelByID :one
SELECT * FROM duels WHERE id = $1;

-- name: GetDuels :many
SELECT * FROM duels
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetVisibleDuels :many
SELECT * FROM duels
WHERE status IN ('waiting_for_opponent', 'in_progress')
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetVisibleDuelsCount :one
SELECT COUNT(*) FROM duels
WHERE status IN ('waiting_for_opponent', 'in_progress');

-- name: GetTopDuels :many
SELECT * FROM duels
WHERE status = 'completed'
ORDER BY created_at DESC, total_stake_value DESC
LIMIT $1 OFFSET $2;

-- name: CreateDuel :one
INSERT INTO duels (is_private, max_players, max_gifts)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateDuelParticipant :one
INSERT INTO duel_participants (duel_id, telegram_user_id, is_creator)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateDuelStake :one
INSERT INTO duel_stakes (duel_id, telegram_user_id, gift_id, stake_value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateDuelRound :one
INSERT INTO duel_rounds (duel_id, round_number)
VALUES ($1, $2)
RETURNING *;

-- name: CreateDuelRoll :one
INSERT INTO duel_rolls (duel_id, round_number, telegram_user_id, dice_value, rolled_at, is_auto_rolled)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;