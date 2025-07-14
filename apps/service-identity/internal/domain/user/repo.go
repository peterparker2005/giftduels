package user

import "context"

type Repository interface {
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	GetUsersByTelegramIDs(ctx context.Context, telegramUserIDs []int64) ([]*User, error)
	UpsertUser(ctx context.Context, user *CreateUserParams) (*User, bool, error)
}
