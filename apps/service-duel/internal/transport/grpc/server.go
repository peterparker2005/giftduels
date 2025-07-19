package grpc

import (
	"net"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/transport/grpc/errors"
	"github.com/peterparker2005/giftduels/packages/grpc-go/interceptors"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
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
	publicHandler duelv1.DuelPublicServiceServer,
	privateHandler duelv1.DuelPrivateServiceServer,
	log *logger.Logger,
) *Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.TelegramIDCtxInterceptor(),
			interceptors.CorrelationInterceptorUnary(),
			interceptors.VersionInterceptorUnary(),
			interceptors.RecoveryInterceptor(log),
			grpcerrors.ErrorMappingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			interceptors.TelegramIDStreamInterceptor(),
			interceptors.CorrelationInterceptorStream(),
			interceptors.VersionInterceptorStream(),
			interceptors.RecoveryInterceptorStream(log),
		),
	}

	s := grpc.NewServer(opts...)

	duelv1.RegisterDuelPublicServiceServer(s, publicHandler)
	duelv1.RegisterDuelPrivateServiceServer(s, privateHandler)

	hs := health.NewServer()
	hs.SetServingStatus("duel", healthpb.HealthCheckResponse_SERVING)
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

func (s *Server) Start() error {
	go func() {
		s.log.Info("Starting gRPC server", zap.String("addr", s.listener.Addr().String()))
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.log.Error("gRPC server stopped with error", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) Stop() error {
	s.log.Info("Stopping gRPC server")
	s.grpcServer.GracefulStop()
	return nil
}
