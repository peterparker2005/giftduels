package builder

import (
	errorsv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/errors/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// BuildError создает gRPC ошибку с указанным grpc-кодом, базовым сообщением и деталью ошибки.
func BuildError(grpcCode codes.Code, baseMessage string, detail *errorsv1.ErrorDetail) error {
	st := status.New(grpcCode, baseMessage)
	anyDetail, err := anypb.New(detail)
	if err != nil {
		// Если не удалось обернуть detail, возвращаем статус без деталей.
		return st.Err()
	}
	st, err = st.WithDetails(anyDetail)
	if err != nil {
		return st.Err()
	}
	return st.Err()
}

// BuildValidationError создает ошибку валидации с указанными параметрами.
// Поле errorDetail.field заполняется только для валидационных ошибок.
func BuildValidationError(field, message string, code errorsv1.ErrorCode) error {
	detail := &errorsv1.ErrorDetail{
		Code:    code,
		Message: message,
		Field:   field,
	}
	return BuildError(codes.InvalidArgument, "validation failed", detail)
}
