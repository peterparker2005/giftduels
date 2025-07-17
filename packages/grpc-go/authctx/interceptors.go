package authctx

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type wrappedStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func TelegramIDStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 1) Извлекаем метадату из исходного контекста
		ctx := ss.Context()
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-telegram-user-id"); len(vals) == 1 {
				if id, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
					// 2) Кладём userID в контекст
					ctx = context.WithValue(ctx, TelegramUserIDKey, id)
				}
			}
		}
		// 3) Оборачиваем ServerStream, чтобы Context() возвращал новый ctx
		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		// 4) Вызываем обработчик с обёрнутым стримом
		return handler(srv, wrapped)
	}
}

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
