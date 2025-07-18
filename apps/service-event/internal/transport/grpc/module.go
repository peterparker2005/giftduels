package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/apps/service-event/internal/config"
	grpchandlers "github.com/peterparker2005/giftduels/apps/service-event/internal/transport/grpc/handlers"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(
		func(cfg *config.Config) (net.Listener, error) {
			return NewListener(&cfg.GRPC.Event)
		},
		grpchandlers.NewEventPublicHandler,
		NewGRPCServer,
	),

	fx.Invoke(func(lc fx.Lifecycle, s *Server) {
		lc.Append(fx.Hook{
			OnStart: s.Start,
			OnStop:  s.Stop,
		})
	}),
)
