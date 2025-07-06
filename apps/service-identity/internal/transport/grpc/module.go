package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		func(cfg *config.Config) net.Listener {
			listener, err := NewListener(&cfg.GRPC.Identity)
			if err != nil {
				panic(err)
			}
			return listener
		},
		func(cfg *config.Config, logger *logger.Logger) token.TokenService {
			return token.NewJWTService(&cfg.JWT, logger.Zap())
		},
		NewRecoveryInterceptor,
		NewVersionInterceptors,
		NewIdentityPublicHandler,
		NewGRPCServer,
	),
	fx.Invoke(
		func(lc fx.Lifecycle, s *Server) {
			lc.Append(fx.Hook{
				OnStart: s.Start,
				OnStop:  s.Stop,
			})
		},
	),
)
