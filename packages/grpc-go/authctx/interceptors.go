package authctx

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TelegramIDCtxInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-telegram-user-id"); len(vals) == 1 {
				if id, errConv := strconv.ParseInt(vals[0], 10, 64); errConv == nil {
					ctx = context.WithValue(ctx, TelegramUserIDKey, id)
				}
			}
		}
		return handler(ctx, req)
	}
}
