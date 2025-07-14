package grpc

import (
	"net"

	envoyauthv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	authctx "github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
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
	recoverInterceptor grpc.UnaryServerInterceptor,
	versionUnary []grpc.UnaryServerInterceptor,
	versionStream []grpc.StreamServerInterceptor,
	publicHandler identityv1.IdentityPublicServiceServer,
	privateHandler identityv1.IdentityPrivateServiceServer,
	envoyHandler envoyauthv3.AuthorizationServer,
	log *logger.Logger,
) *Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			append(versionUnary, recoverInterceptor, authctx.TelegramIDCtxInterceptor())...),
		grpc.ChainStreamInterceptor(versionStream...),
	}

	s := grpc.NewServer(opts...)

	identityv1.RegisterIdentityPublicServiceServer(s, publicHandler)
	identityv1.RegisterIdentityPrivateServiceServer(s, privateHandler)
	envoyauthv3.RegisterAuthorizationServer(s, envoyHandler)
	hs := health.NewServer()
	hs.SetServingStatus("identity", healthpb.HealthCheckResponse_SERVING)
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
