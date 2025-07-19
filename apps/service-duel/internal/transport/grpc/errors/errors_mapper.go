package grpcerrors

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/query"
	pkgerrors "github.com/peterparker2005/giftduels/packages/errors"
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// MapError converts any error from services/repositories
// to a gRPC status with details and trace-id from ctx.
//
//nolint:funlen // map error is complex
func MapError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// --- Query errors ---
	switch {
	case query.IsGetGifts(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to fetch gifts"),
			pkgerrors.WithContext(ctx),
		)
	case query.IsGiftNotFoundInResponse(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("gift not found in response"),
			pkgerrors.WithContext(ctx),
		)
	case query.IsParseGiftPrice(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to parse gift price"),
			pkgerrors.WithContext(ctx),
		)
	case query.IsBuildStakedGift(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to build staked gift"),
			pkgerrors.WithContext(ctx),
		)
	case query.IsGetUsers(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to fetch users"),
			pkgerrors.WithContext(ctx),
		)
	}

	// --- Command errors ---
	switch {
	case command.IsCreateDuel(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to create duel"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsJoinDuel(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to join duel"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsDuelNotFound(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.NotFound),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_NOT_FOUND),
			pkgerrors.WithMessage("duel not found"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsDuelFull(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.ResourceExhausted),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_DUEL_FULL),
			pkgerrors.WithMessage("duel is full"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsAlreadyJoined(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.AlreadyExists),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_ALREADY_EXISTS),
			pkgerrors.WithMessage("already joined this duel"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsStakeOutOfRange(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.InvalidArgument),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL),
			pkgerrors.WithMessage("stake is out of allowed entry range"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsGiftStakingFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to stake gift"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsInvalidGiftPrice(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.InvalidArgument),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL),
			pkgerrors.WithMessage("invalid gift price"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsAutoRoll(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to auto roll"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsRollDiceFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to roll dice"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsNoCurrentRound(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.InvalidArgument),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL),
			pkgerrors.WithMessage("no current round"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsRoundEvaluationFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to evaluate round"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsStartNewRoundFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to start new round"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsCompleteDuelFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to complete duel"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsSendDuelCompletedMessageFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to send duel completed message"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsTransactionFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_DATABASE_ERROR),
			pkgerrors.WithMessage("transaction failed"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsDatabaseOperation(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_DATABASE_ERROR),
			pkgerrors.WithMessage("database operation failed"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsInvalidParticipant(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.InvalidArgument),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL),
			pkgerrors.WithMessage("invalid participant"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsInvalidStake(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.InvalidArgument),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL),
			pkgerrors.WithMessage("invalid stake"),
			pkgerrors.WithContext(ctx),
		)
	case command.IsPublishEventFailed(err):
		return pkgerrors.NewError(
			pkgerrors.WithGRPCCode(codes.Internal),
			pkgerrors.WithErrorCode(errorsv1.ErrorCode_ERROR_CODE_INTERNAL),
			pkgerrors.WithMessage("failed to publish event"),
			pkgerrors.WithContext(ctx),
		)
	}

	// All other errors are Internal with a wrapper.
	return pkgerrors.Wrap(ctx, err)
}

// ErrorMappingInterceptor maps errors to gRPC statuses.
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
