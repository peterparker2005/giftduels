package grpcerrors

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/service/payment"
	"github.com/peterparker2005/giftduels/packages/errors"
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	"google.golang.org/grpc/codes"
)

func MapError(ctx context.Context, err error) error {
	if payment.IsInsufficientBalance(err) {
		return errors.NewError(
			errors.WithGRPCCode(codes.ResourceExhausted),
			errors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INSUFFICIENT_BALANCE),
			errors.WithMessage("insufficient balance"),
			errors.WithContext(ctx),
		)
	}
	return errors.Wrap(ctx, err)
}
