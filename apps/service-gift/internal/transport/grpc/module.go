package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		func(cfg *config.Config) (net.Listener, error) {
			return NewListener(&cfg.GRPC.Gift)
		},
		NewRecoveryInterceptor,
		NewVersionInterceptors,
		NewGiftPublicHandler,
		NewGiftPrivateHandler,
		NewGRPCServer,
	),

	fx.Invoke(func(lc fx.Lifecycle, s *Server) {
		lc.Append(fx.Hook{
			OnStart: s.Start,
			OnStop:  s.Stop,
		})
	}),
)
