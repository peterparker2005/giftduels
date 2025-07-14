package grpc

import (
	"context"

	duelService "github.com/peterparker2005/giftduels/apps/service-duel/internal/service/duel"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
)

type DuelPrivateHandler struct {
	duelv1.UnimplementedDuelPrivateServiceServer

	duelService *duelService.Service
}

func NewDuelPrivateHandler(duelService *duelService.Service) duelv1.DuelPrivateServiceServer {
	return &DuelPrivateHandler{duelService: duelService}
}

func (h *DuelPrivateHandler) FindDuelByGiftID(ctx context.Context, req *duelv1.FindDuelByGiftIDRequest) (*duelv1.FindDuelByGiftIDResponse, error) {
	duelID, err := h.duelService.FindDuelByGiftID(ctx, req.GetGiftId().GetValue())
	if err != nil {
		return nil, errors.NewInternalError("failed to find duel by gift id")
	}
	return &duelv1.FindDuelByGiftIDResponse{DuelId: &sharedv1.DuelId{Value: duelID.String()}}, nil
}
