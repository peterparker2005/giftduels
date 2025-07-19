package grpc

import (
	"context"
	"net"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	grpchandlers "github.com/peterparker2005/giftduels/apps/service-identity/internal/transport/grpc/handlers"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(
		func(cfg *config.Config) net.Listener {
			listener, err := NewListener(&cfg.GRPC.Identity)
			if err != nil {
				panic(err)
			}
			return listener
		},
		grpchandlers.NewIdentityPublicHandler,
		grpchandlers.NewIdentityPrivateHandler,
		grpchandlers.NewIdentityEnvoyHandler,
		NewGRPCServer,
	),
	fx.Invoke(
		func(lc fx.Lifecycle, s *Server) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return s.Start()
				},
				OnStop: func(_ context.Context) error {
					return s.Stop()
				},
			})
		},
	),
)
