package grpc

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewListener,            // net.Listener
		NewRecoveryInterceptor, // grpc.UnaryServerInterceptor
		NewVersionInterceptors, // []grpc.StreamServerInterceptor, []grpc.UnaryServerInterceptor
		NewGiftPublicHandler,   // pb.IdentityPublicServiceServer
		NewGiftPrivateHandler,  // pb.IdentityPrivateServiceServer
		NewServer,              // (*Server).Ctor
	),
	fx.Invoke(
		registerHealthCheck,
		registerReflection,
		startServer,
	),
)
