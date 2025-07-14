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

-- name: GetDuelParticipants :many
SELECT * FROM duel_participants WHERE duel_id = $1;

-- name: GetDuelStakes :many
SELECT * FROM duel_stakes WHERE duel_id = $1;

-- name: GetDuelRounds :many
SELECT * FROM duel_rounds WHERE duel_id = $1 ORDER BY round_number;

-- name: GetDuelRolls :many
SELECT * FROM duel_rolls WHERE duel_id = $1 ORDER BY round_number, rolled_at;

-- name: CreateDuel :one
INSERT INTO duels (is_private, max_players, max_gifts)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateParticipant :one
INSERT INTO duel_participants (duel_id, telegram_user_id, is_creator)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateStake :one
INSERT INTO duel_stakes (duel_id, telegram_user_id, gift_id, stake_value)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateRound :one
INSERT INTO duel_rounds (duel_id, round_number)
VALUES ($1, $2)
RETURNING *;

-- name: CreateRoll :one
INSERT INTO duel_rolls (duel_id, round_number, telegram_user_id, dice_value, rolled_at, is_auto_rolled)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: FindDueDuels :many
SELECT id FROM duels
WHERE status = 'in_progress'
  AND next_roll_deadline <= $1;

-- name: UpdateNextRollDeadline :exec
UPDATE duels
  SET next_roll_deadline = $2
WHERE id = $1;

-- name: UpdateDuelStatus :exec
UPDATE duels
  SET status = $2,
      winner_telegram_user_id = $3,
      completed_at = $4
WHERE id = $1;

-- name: FindDuelByGiftID :one
SELECT d.id
FROM duels d
JOIN duel_stakes s ON s.duel_id = d.id
WHERE s.gift_id = $1
ORDER BY d.created_at DESC
LIMIT 1;
