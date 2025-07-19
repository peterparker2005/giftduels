package interceptors

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor recovers from panics and logs them.
func RecoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(
		grpc_recovery.WithRecoveryHandler(recoveryHandler(log)),
	)
}

func RecoveryInterceptorStream(log *logger.Logger) grpc.StreamServerInterceptor {
	return grpc_recovery.StreamServerInterceptor(
		grpc_recovery.WithRecoveryHandler(recoveryHandler(log)),
	)
}

func recoveryHandler(log *logger.Logger) grpc_recovery.RecoveryHandlerFunc {
	return func(p interface{}) error {
		log.Error("panic", zap.Any("panic", p))
		return status.Errorf(codes.Internal, "internal error")
	}
}
