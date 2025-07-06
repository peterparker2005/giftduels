package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(
		// Network listener
		func(cfg *config.Config) net.Listener {
			listener, err := NewListener(&cfg.GRPC.Identity)
			if err != nil {
				panic(err)
			}
			return listener
		},
		// Service dependencies
		func(cfg *config.Config, logger *logger.Logger) token.TokenService {
			return token.NewJWTService(&cfg.JWT, logger.Zap())
		},
		func(logger *logger.Logger) *zap.Logger {
			return logger.Zap()
		},
		// gRPC components
		NewRecoveryInterceptor,   // grpc.UnaryServerInterceptor
		NewVersionInterceptors,   // []grpc.StreamServerInterceptor, []grpc.UnaryServerInterceptor
		NewIdentityPublicHandler, // *IdentityPublicHandler
		NewServer,                // *grpc.Server
	),
	fx.Invoke(
		registerHealthCheck,
		registerReflection,
		startServer,
	),
)
