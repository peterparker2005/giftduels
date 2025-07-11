package grpc

import (
	"context"
	"net"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	authctx "github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	cfg        *config.Config
	log        *logger.Logger
}

func NewGRPCServer(
	cfg *config.Config,
	listener net.Listener,
	recover grpc.UnaryServerInterceptor,
	versionUnary []grpc.UnaryServerInterceptor,
	versionStream []grpc.StreamServerInterceptor,
	publicHandler giftv1.GiftPublicServiceServer,
	privateHandler giftv1.GiftPrivateServiceServer,
	log *logger.Logger,
) *Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(append(versionUnary, recover, authctx.TelegramIDCtxInterceptor())...),
		grpc.ChainStreamInterceptor(versionStream...),
	}

	s := grpc.NewServer(opts...)

	giftv1.RegisterGiftPublicServiceServer(s, publicHandler)
	giftv1.RegisterGiftPrivateServiceServer(s, privateHandler)

	hs := health.NewServer()
	hs.SetServingStatus("gift", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hs)

	if cfg.Environment != "production" {
		reflection.Register(s)
	}

	return &Server{
		grpcServer: s,
		listener:   listener,
		cfg:        cfg,
		log:        log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		s.log.Info("Starting gRPC server", zap.String("addr", s.listener.Addr().String()))
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.log.Error("gRPC server stopped with error", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Stopping gRPC server")
	s.grpcServer.GracefulStop()
	return nil
}
