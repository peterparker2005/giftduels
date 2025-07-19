package authctx

import (
	"context"

	"github.com/peterparker2005/giftduels/packages/errors"
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	"google.golang.org/grpc/codes"
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
		return 0, errors.NewError(
			errors.WithGRPCCode(codes.Unauthenticated),
			errors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_UNAUTHORIZED),
			errors.WithMessage("telegram user id not found"),
		)
	}
	return id, nil
}
