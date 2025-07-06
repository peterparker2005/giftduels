package version

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryInterceptor прикрепляет заголовок x-service-version ко всем ответам
func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// шлём метаданные до основного handler-а
		_ = grpc.SendHeader(ctx, metadata.Pairs("x-service-version", Version))
		return handler(ctx, req)
	}
}

// StreamInterceptor для стримов
func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		_ = ss.SendHeader(metadata.Pairs("x-service-version", Version))
		return handler(srv, ss)
	}
}
