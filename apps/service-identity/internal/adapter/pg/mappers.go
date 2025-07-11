package pg

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/domain/user"
)

func UserToDomain(u sqlc.User) *user.User {
	return &user.User{
		ID:              u.ID.String(),
		TelegramID:      u.TelegramID,
		Username:        u.Username.String,
		FirstName:       u.FirstName,
		LastName:        u.LastName.String,
		LanguageCode:    u.LanguageCode.String,
		AllowsWriteToPm: u.AllowsWriteToPm.Bool,
		IsPremium:       u.IsPremium.Bool,
		PhotoUrl:        u.PhotoUrl.String,
		CreatedAt:       u.CreatedAt.Time,
		UpdatedAt:       u.UpdatedAt.Time,
	}
}

func UpsertUserParamsToSQLC(params *user.CreateUserParams) sqlc.UpsertUserParams {
	return sqlc.UpsertUserParams{
		TelegramID:      params.TelegramID,
		Username:        pgtype.Text{String: params.Username, Valid: params.Username != ""},
		FirstName:       params.FirstName,
		LastName:        pgtype.Text{String: params.LastName, Valid: params.LastName != ""},
		LanguageCode:    pgtype.Text{String: params.LanguageCode, Valid: params.LanguageCode != ""},
		AllowsWriteToPm: pgtype.Bool{Bool: params.AllowsWriteToPm, Valid: true},
		IsPremium:       pgtype.Bool{Bool: params.IsPremium, Valid: true},
		PhotoUrl:        pgtype.Text{String: params.PhotoUrl, Valid: params.PhotoUrl != ""},
	}
}
