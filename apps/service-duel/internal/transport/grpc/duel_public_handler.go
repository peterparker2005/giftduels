package grpc

import (
	"context"

	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelService "github.com/peterparker2005/giftduels/apps/service-duel/internal/service/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type duelPublicHandler struct {
	duelv1.UnimplementedDuelPublicServiceServer

	logger      *logger.Logger
	duelService *duelService.DuelService
}

// NewDuelPublicHandler создает новый GRPC handler
func NewDuelPublicHandler(logger *logger.Logger, duelService *duelService.DuelService) duelv1.DuelPublicServiceServer {
	return &duelPublicHandler{
		logger:      logger,
		duelService: duelService,
	}
}

func (h *duelPublicHandler) GetDuelList(ctx context.Context, req *duelv1.GetDuelListRequest) (*duelv1.GetDuelListResponse, error) {
	pageRequest := shared.NewPageRequest(req.GetPageRequest().GetPage(), req.GetPageRequest().GetPageSize())

	response, err := h.duelService.GetDuelList(ctx, pageRequest)
	if err != nil {
		return nil, err
	}

	// Map duels to protobuf
	protoDuels := make([]*duelv1.Duel, len(response.Duels))
	for i, duel := range response.Duels {
		protoDuels[i] = mapDuel(duel)
	}

	return &duelv1.GetDuelListResponse{
		Duels: protoDuels,
	}, nil
}

func (h *duelPublicHandler) CreateDuel(ctx context.Context, req *duelv1.CreateDuelRequest) (*duelv1.CreateDuelResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	params, err := mapDuelParamsFromProto(req.Params)
	if err != nil {
		return nil, err
	}

	// Map stakes
	stakes := make([]duelDomain.Stake, len(req.Stakes))
	for i, stake := range req.Stakes {
		stakes[i] = duelDomain.Stake{
			GiftID: stake.GiftId.Value,
		}
	}

	createParams := duelService.CreateDuelParams{
		Params: params,
		Participants: []duelDomain.Participant{
			{
				// FIXME: map correctly
				TelegramUserID: duelDomain.TelegramUserID(telegramUserID),
				IsCreator:      true,
			},
		},
		Stakes: stakes,
	}

	duelID, err := h.duelService.CreateDuel(ctx, telegramUserID, createParams)
	if err != nil {
		return nil, err
	}

	return &duelv1.CreateDuelResponse{
		DuelId: &sharedv1.DuelId{Value: duelID.String()},
	}, nil
}
