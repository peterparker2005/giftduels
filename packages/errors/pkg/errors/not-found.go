package errors

import (
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/peterparker2005/giftduels/packages/errors/internal/builder"
)

// NewNotFoundError создаёт ошибку "не найдено"
func NewNotFoundError(message string) error {
	detail := &errorsv1.ErrorDetail{
		Code:    errorsv1.ErrorCode_ERROR_CODE_NOT_FOUND,
		Message: message,
	}
	return builder.BuildError(codes.NotFound, "not found", detail)
}

// IsNotFound проверяет, является ли ошибка ошибкой "не найдено"
func IsNotFound(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.NotFound
}
