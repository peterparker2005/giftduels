package grpc

import (
	"context"

	"github.com/ccoveille/go-safecast"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/proto"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/query"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"go.uber.org/zap"
)

type duelPublicHandler struct {
	duelv1.UnimplementedDuelPublicServiceServer

	logger              *logger.Logger
	duelCreateCommand   *command.DuelCreateCommand
	duelAutoRollCommand *command.DuelAutoRollCommand
	duelQueryService    *query.DuelQueryService
}

// NewDuelPublicHandler создает новый GRPC handler.
func NewDuelPublicHandler(
	logger *logger.Logger,
	duelCreateCommand *command.DuelCreateCommand,
	duelAutoRollCommand *command.DuelAutoRollCommand,
	duelQueryService *query.DuelQueryService,
) duelv1.DuelPublicServiceServer {
	return &duelPublicHandler{
		logger:              logger,
		duelCreateCommand:   duelCreateCommand,
		duelAutoRollCommand: duelAutoRollCommand,
		duelQueryService:    duelQueryService,
	}
}

func (h *duelPublicHandler) GetDuelList(
	ctx context.Context,
	req *duelv1.GetDuelListRequest,
) (*duelv1.GetDuelListResponse, error) {
	pageRequest := shared.NewPageRequest(
		req.GetPageRequest().GetPage(),
		req.GetPageRequest().GetPageSize(),
	)

	response, err := h.duelQueryService.GetDuelList(ctx, pageRequest)
	if err != nil {
		h.logger.Error("failed to get duel list", zap.Error(err))
		return nil, errors.NewInternalError("failed to get duel list")
	}

	// Map duels to protobuf
	protoDuels := make([]*duelv1.Duel, len(response.Duels))
	for i, duel := range response.Duels {
		protoDuels[i], err = proto.MapDuel(duel)
		if err != nil {
			return nil, errors.NewInternalError("failed to map duel")
		}
	}

	total, err := safecast.ToInt32(response.Total)
	if err != nil {
		return nil, errors.NewInternalError("failed to cast total")
	}

	return &duelv1.GetDuelListResponse{
		Duels: protoDuels,
		Pagination: &sharedv1.PageResponse{
			TotalPages: pageRequest.TotalPages(total),
			Total:      total,
			Page:       pageRequest.Page(),
			PageSize:   pageRequest.PageSize(),
		},
	}, nil
}

func (h *duelPublicHandler) CreateDuel(
	ctx context.Context,
	req *duelv1.CreateDuelRequest,
) (*duelv1.CreateDuelResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	params, err := proto.MapDuelParamsFromProto(req.GetParams())
	if err != nil {
		return nil, errors.NewInternalError("failed to create duel")
	}

	// Map stakes
	stakes := make([]duelDomain.Stake, len(req.GetStakes()))
	for i, stake := range req.GetStakes() {
		giftID := stake.GetGiftId().GetValue()
		h.logger.Info("processing stake",
			zap.String("giftID", giftID),
			zap.Int("stakeIndex", i))

		stakes[i] = duelDomain.Stake{
			Gift: duelDomain.NewStakedGift(
				giftID,
				"",
				"",
				nil,
			),
		}
	}

	createParams := command.CreateDuelParams{
		Params: params,
		Stakes: stakes,
	}

	duelID, err := h.duelCreateCommand.Execute(ctx, telegramUserID, createParams)
	if err != nil {
		h.logger.Error("failed to create duel",
			zap.Int64("telegramUserID", telegramUserID),
			zap.Int("stakesCount", len(req.GetStakes())),
			zap.Error(err))
		return nil, err
	}

	return &duelv1.CreateDuelResponse{
		DuelId: &sharedv1.DuelId{Value: duelID.String()},
	}, nil
}

func (h *duelPublicHandler) GetDuel(
	ctx context.Context,
	req *duelv1.GetDuelRequest,
) (*duelv1.GetDuelResponse, error) {
	duelIDStr := req.GetId().GetValue()
	duelID, err := duelDomain.NewID(duelIDStr)
	if err != nil {
		return nil, errors.NewInternalError("failed to create duel ID")
	}
	duel, err := h.duelQueryService.GetDuelByID(ctx, duelID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get duel")
	}

	protoDuel, err := proto.MapDuel(duel)
	if err != nil {
		return nil, errors.NewInternalError("failed to map duel")
	}

	return &duelv1.GetDuelResponse{Duel: protoDuel}, nil
}
