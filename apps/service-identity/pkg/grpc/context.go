package grpc

import (
	"context"

	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
)

func TelegramUserID(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(TelegramUserIDKey).(int64)
	if !ok {
		return 0, errors.NewUnauthorizedError("telegram user id not found")
	}
	return id, nil
}
