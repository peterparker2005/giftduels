package grpc

import (
	"context"
	"net"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	pb "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
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
	publicHandler pb.GiftPublicServiceServer,
	privateHandler pb.GiftPrivateServiceServer,
) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(append(versionUnary, recover)...),
		grpc.ChainStreamInterceptor(versionStream...),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterGiftPublicServiceServer(srv, publicHandler)
	pb.RegisterGiftPrivateServiceServer(srv, privateHandler)

	return srv
}

func registerHealthCheck(srv *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("gift", healthpb.HealthCheckResponse_SERVING)
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
