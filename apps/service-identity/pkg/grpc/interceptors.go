package grpc

import (
	"context"
	"strings"

	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func AuthInterceptor(client identityv1.IdentityPublicServiceClient) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Если метод не защищён — просто пропускаем
		needAuth := protectedMethods[info.FullMethod]
		if !needAuth {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.NewUnauthorizedError("missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, errors.NewUnauthorizedError("missing authorization header")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == "" {
			return nil, errors.NewUnauthorizedError("invalid bearer token")
		}

		res, err := client.ValidateToken(ctx, &identityv1.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			return nil, errors.NewUnauthorizedError("invalid or expired token")
		}

		ctx = context.WithValue(ctx, TelegramUserIDKey, res.TelegramUserId)

		return handler(ctx, req)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func StreamServerInterceptor(client identityv1.IdentityPublicServiceClient) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Если метод не требует авторизации — пропускаем
		needAuth := protectedMethods[info.FullMethod]
		if !needAuth {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.NewUnauthorizedError("missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return errors.NewUnauthorizedError("missing authorization header")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == "" {
			return errors.NewUnauthorizedError("invalid bearer token")
		}

		res, err := client.ValidateToken(ss.Context(), &identityv1.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			return errors.NewUnauthorizedError("invalid or expired token")
		}

		ctx := context.WithValue(ss.Context(), TelegramUserIDKey, res.TelegramUserId)

		// Обернём оригинальный ServerStream, чтобы вернуть модифицированный контекст
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}
