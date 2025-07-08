package authctx

import (
	"context"

	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
)

type contextKey string

const (
	TelegramUserIDKey = contextKey("x-telegram-user-id")
)

func (k contextKey) String() string {
	return string(k)
}

func TelegramUserID(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(TelegramUserIDKey).(int64)
	if !ok {
		return 0, errors.NewUnauthorizedError("telegram user id not found")
	}
	return id, nil
}
