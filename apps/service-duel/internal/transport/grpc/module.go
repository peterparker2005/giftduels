package grpc

import (
	"context"
	"net"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	grpchandlers "github.com/peterparker2005/giftduels/apps/service-duel/internal/transport/grpc/handlers"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(
		func(cfg *config.Config) (net.Listener, error) {
			return NewListener(&cfg.GRPC.Duel)
		},
		grpchandlers.NewDuelPublicHandler,
		grpchandlers.NewDuelPrivateHandler,
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
