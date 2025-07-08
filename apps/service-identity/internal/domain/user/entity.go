package user

import (
	"time"
)

type User struct {
	ID              string
	TelegramID      int64
	Username        string
	FirstName       string
	LastName        string
	PhotoUrl        string
	LanguageCode    string
	AllowsWriteToPm bool
	IsPremium       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
