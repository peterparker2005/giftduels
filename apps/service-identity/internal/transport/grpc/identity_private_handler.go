package grpc

import (
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/token"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
)

var _ identityv1.IdentityPrivateServiceServer = (*IdentityPrivateHandler)(nil)

type IdentityPrivateHandler struct {
	identityv1.UnimplementedIdentityPrivateServiceServer

	tokenSvc token.Service
	logger   *logger.Logger
}

func NewIdentityPrivateHandler(ts token.Service, lg *logger.Logger) identityv1.IdentityPrivateServiceServer {
	return &IdentityPrivateHandler{tokenSvc: ts, logger: lg}
}
