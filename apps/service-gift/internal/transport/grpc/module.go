package grpc

import (
	"context"
	"net"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
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
			OnStart: func(_ context.Context) error {
				return s.Start()
			},
			OnStop: func(_ context.Context) error {
				return s.Stop()
			},
		})
	}),
)
