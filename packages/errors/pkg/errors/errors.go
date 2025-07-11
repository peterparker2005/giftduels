package errors

import (
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/peterparker2005/giftduels/packages/errors/internal/builder"
)

func NewValidationError(field, message string) error {
	return builder.BuildValidationError(field, message, errorsv1.ErrorCode_ERROR_CODE_VALIDATION_GENERAL)
}

// NewRequiredFieldError создает ошибку валидации для указанного поля.
func NewRequiredFieldError(field, message string) error {
	return builder.BuildValidationError(field, message, errorsv1.ErrorCode_ERROR_CODE_REQUIRED_FIELD)
}

// NewUnauthorizedError возвращает ошибку, сигнализирующую об отсутствии аутентификации.
func NewUnauthorizedError(message string) error {
	detail := &errorsv1.ErrorDetail{
		Code:    errorsv1.ErrorCode_ERROR_CODE_UNAUTHORIZED,
		Message: message,
	}
	return builder.BuildError(codes.Unauthenticated, "unauthorized", detail)
}

// NewInternalError создает внутреннюю ошибку с базовым сообщением.
func NewInternalError(message string) error {
	detail := &errorsv1.ErrorDetail{
		Code:    errorsv1.ErrorCode_ERROR_CODE_INTERNAL,
		Message: message,
	}
	return builder.BuildError(codes.Internal, "internal error", detail)
}

// WrapError оборачивает существующую ошибку в gRPC статус, если это необходимо.
func WrapError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return status.Error(codes.Internal, err.Error())
	}
	return st.Err()
}

func NewGiftNotOwnedError(message string) error {
	detail := &errorsv1.ErrorDetail{
		Code:    errorsv1.ErrorCode_ERROR_CODE_GIFT_NOT_OWNED,
		Message: message,
	}
	return builder.BuildError(codes.PermissionDenied, "gift not owned", detail)
}

func NewInsufficientTonError(message string) error {
	detail := &errorsv1.ErrorDetail{
		Code:    errorsv1.ErrorCode_ERROR_CODE_INSUFFICIENT_TON,
		Message: message,
	}
	return builder.BuildError(codes.FailedPrecondition, "insufficient ton", detail)
}

func IsInsufficientTon(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.FailedPrecondition
}
