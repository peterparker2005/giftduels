package user

import "context"

type CreateUserParams struct {
	TelegramID      int64
	Username        string
	FirstName       string
	LastName        string
	PhotoUrl        string
	LanguageCode    string
	AllowsWriteToPm bool
	IsPremium       bool
}

type UserRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	CreateUser(ctx context.Context, params CreateUserParams) (*User, error)
	UpsertUser(ctx context.Context, params CreateUserParams) (*User, error)
}
