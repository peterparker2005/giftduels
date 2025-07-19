package grpchandlers

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/query"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
)

type DuelPrivateHandler struct {
	duelv1.UnimplementedDuelPrivateServiceServer

	duelCreateCommand   *command.DuelCreateCommand
	duelAutoRollCommand *command.DuelAutoRollCommand
	duelQueryService    *query.DuelQueryService
}

func NewDuelPrivateHandler(
	duelCreateCommand *command.DuelCreateCommand,
	duelAutoRollCommand *command.DuelAutoRollCommand,
	duelQueryService *query.DuelQueryService,
) duelv1.DuelPrivateServiceServer {
	return &DuelPrivateHandler{
		duelCreateCommand:   duelCreateCommand,
		duelAutoRollCommand: duelAutoRollCommand,
		duelQueryService:    duelQueryService,
	}
}

func (h *DuelPrivateHandler) FindDuelByGiftID(
	ctx context.Context,
	req *duelv1.FindDuelByGiftIDRequest,
) (*duelv1.FindDuelByGiftIDResponse, error) {
	duelID, err := h.duelQueryService.FindDuelByGiftID(ctx, req.GetGiftId().GetValue())
	if err != nil {
		return nil, err
	}
	return &duelv1.FindDuelByGiftIDResponse{DuelId: &sharedv1.DuelId{Value: duelID.String()}}, nil
}
