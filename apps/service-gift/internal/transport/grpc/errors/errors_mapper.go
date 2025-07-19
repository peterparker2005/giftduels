package grpcerrors

import (
	"context"

	"github.com/peterparker2005/giftduels/packages/errors"
	"google.golang.org/grpc"
)

func MapError(ctx context.Context, err error) error {
	return errors.Wrap(ctx, err)
}

func ErrorMappingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, MapError(ctx, err)
		}
		return resp, nil
	}
}
