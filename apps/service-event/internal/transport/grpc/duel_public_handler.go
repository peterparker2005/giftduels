package grpc

import (
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
)

type duelPublicHandler struct {
	duelv1.DuelPublicServiceServer

	logger *logger.Logger
}

// NewGiftPublicHandler создает новый GRPC handler.
func NewDuelPublicHandler(logger *logger.Logger) duelv1.DuelPublicServiceServer {
	return &duelPublicHandler{
		logger: logger,
	}
}
