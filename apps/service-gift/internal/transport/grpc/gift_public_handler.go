package grpc

import (
	"context"

	domainGift "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"go.uber.org/zap"
)

type giftPublicHandler struct {
	giftv1.GiftPublicServiceServer

	// зависимость от сервисного слоя
	giftService *gift.Service
	logger      *logger.Logger
}

// NewGiftPublicHandler создает новый GRPC handler
func NewGiftPublicHandler(giftService *gift.Service, logger *logger.Logger) giftv1.GiftPublicServiceServer {
	return &giftPublicHandler{
		giftService: giftService,
		logger:      logger,
	}
}

func (h *giftPublicHandler) GetGift(ctx context.Context, req *giftv1.GetGiftRequest) (*giftv1.GetGiftResponse, error) {
	g, err := h.giftService.GetGiftByID(ctx, req.GetGiftId().Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.GetGiftResponse{
		Gift: domainGift.DomainGiftToProtoView(g),
	}, nil
}

func (h *giftPublicHandler) GetGifts(ctx context.Context, req *giftv1.GetGiftsRequest) (*giftv1.GetGiftsResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	log := h.logger.With(zap.Int64("telegramUserID", telegramUserID))

	pagination := shared.NewPageRequest(req.GetPagination().GetPage(), req.GetPagination().GetPageSize())
	domainGifts, err := h.giftService.GetUserActiveGifts(ctx, telegramUserID, pagination)
	if err != nil {
		log.Error("Failed to get user active gifts", zap.Error(err))
		return nil, err
	}

	giftViews := make([]*giftv1.GiftView, len(domainGifts.Gifts))
	for i, g := range domainGifts.Gifts {
		giftViews[i] = domainGift.DomainGiftToProtoView(g)
	}

	return &giftv1.GetGiftsResponse{
		Gifts:      giftViews,
		TotalValue: domainGifts.TotalValue,
		Pagination: &sharedv1.PageResponse{
			Page:       pagination.Page(),
			PageSize:   pagination.PageSize(),
			Total:      domainGifts.Total,
			TotalPages: pagination.TotalPages(domainGifts.Total),
		},
	}, nil
}

func (h *giftPublicHandler) ExecuteWithdraw(ctx context.Context, req *giftv1.ExecuteWithdrawRequest) (*giftv1.ExecuteWithdrawResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(req.GetGiftIds()))
	for i, id := range req.GetGiftIds() {
		ids[i] = id.Value
	}
	_, err = h.giftService.ExecuteWithdraw(ctx, telegramUserID, ids)
	if err != nil {
		return nil, err
	}

	// For now, return success response
	// TODO: implement actual withdrawal logic based on method
	return &giftv1.ExecuteWithdrawResponse{
		Success: &sharedv1.SuccessResponse{
			Success: true,
			Message: "Withdrawal request received",
		},
	}, nil
}
