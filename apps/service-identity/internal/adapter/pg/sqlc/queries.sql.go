// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getUserByTelegramID = `-- name: GetUserByTelegramID :one
SELECT id, telegram_id, username, first_name, last_name, photo_url, language_code, allows_write_to_pm, is_premium, created_at, updated_at FROM users WHERE telegram_id = $1 LIMIT 1
`

func (q *Queries) GetUserByTelegramID(ctx context.Context, telegramID int64) (User, error) {
	row := q.db.QueryRow(ctx, getUserByTelegramID, telegramID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.TelegramID,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.PhotoUrl,
		&i.LanguageCode,
		&i.AllowsWriteToPm,
		&i.IsPremium,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUsersByTelegramIDs = `-- name: GetUsersByTelegramIDs :many
SELECT id, telegram_id, username, first_name, last_name, photo_url, language_code, allows_write_to_pm, is_premium, created_at, updated_at FROM users WHERE telegram_id = ANY($1::bigint[])
`

func (q *Queries) GetUsersByTelegramIDs(ctx context.Context, dollar_1 []int64) ([]User, error) {
	rows, err := q.db.Query(ctx, getUsersByTelegramIDs, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.TelegramID,
			&i.Username,
			&i.FirstName,
			&i.LastName,
			&i.PhotoUrl,
			&i.LanguageCode,
			&i.AllowsWriteToPm,
			&i.IsPremium,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertUser = `-- name: UpsertUser :one
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
RETURNING id, telegram_id, username, first_name, last_name, photo_url, language_code, allows_write_to_pm, is_premium, created_at, updated_at
`

type UpsertUserParams struct {
	TelegramID      int64
	Username        pgtype.Text
	FirstName       string
	LastName        pgtype.Text
	PhotoUrl        pgtype.Text
	LanguageCode    pgtype.Text
	AllowsWriteToPm pgtype.Bool
	IsPremium       pgtype.Bool
}

func (q *Queries) UpsertUser(ctx context.Context, arg UpsertUserParams) (User, error) {
	row := q.db.QueryRow(ctx, upsertUser,
		arg.TelegramID,
		arg.Username,
		arg.FirstName,
		arg.LastName,
		arg.PhotoUrl,
		arg.LanguageCode,
		arg.AllowsWriteToPm,
		arg.IsPremium,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.TelegramID,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.PhotoUrl,
		&i.LanguageCode,
		&i.AllowsWriteToPm,
		&i.IsPremium,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
