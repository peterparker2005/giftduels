package grpc

import (
	"context"
	"net"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(
	recover grpc.UnaryServerInterceptor,
	versionUnary []grpc.UnaryServerInterceptor,
	versionStream []grpc.StreamServerInterceptor,
	publicHandler *IdentityPublicHandler,
) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(append(versionUnary, recover)...),
		grpc.ChainStreamInterceptor(versionStream...),
	}
	srv := grpc.NewServer(opts...)
	identityv1.RegisterIdentityPublicServiceServer(srv, publicHandler)

	return srv
}

func registerHealthCheck(srv *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("identity", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)
}

func registerReflection(cfg *config.Config, srv *grpc.Server) {
	if cfg.Environment != "production" {
		reflection.Register(srv)
	}
}

func startServer(
	lc fx.Lifecycle,
	srv *grpc.Server,
	lis net.Listener,
	log *logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.Serve(lis); err != nil {
					log.Error("failed to serve gRPC", zap.Error(err))
				}
			}()
			log.Info("gRPC listening", zap.String("addr", lis.Addr().String()))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping gRPC")
			srv.GracefulStop()
			return nil
		},
	})
}
