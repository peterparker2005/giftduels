package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		func(cfg *config.Config) net.Listener {
			listener, err := NewListener(cfg.GRPC.Payment.Address())
			if err != nil {
				panic(err)
			}
			return listener
		},
		NewRecoveryInterceptor,
		NewVersionInterceptors,
		NewPaymentPublicHandler,
		NewPaymentPrivateHandler,
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
