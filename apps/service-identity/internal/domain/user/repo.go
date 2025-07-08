package user

import "context"

type UserRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	CreateOrUpdate(ctx context.Context, user *CreateUserParams) (*User, bool, error)
}
