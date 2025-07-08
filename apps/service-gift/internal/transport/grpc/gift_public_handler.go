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
		Gift: domainGift.ConvertDomainGiftToProtoView(g),
	}, nil
}

func (h *giftPublicHandler) GetGifts(ctx context.Context, req *giftv1.GetGiftsRequest) (*giftv1.GetGiftsResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	pagination := shared.NewPageRequest(req.GetPagination().GetPage(), req.GetPagination().GetPageSize())
	h.logger.Info("GetGifts", zap.Any("pagination", pagination))
	domainGifts, err := h.giftService.GetUserGifts(ctx, telegramUserID, pagination)
	if err != nil {
		h.logger.Error("GetGifts", zap.Error(err))
		return nil, err
	}

	giftViews := make([]*giftv1.GiftView, len(domainGifts))
	for i, g := range domainGifts {
		h.logger.Info("GetGifts", zap.Any("domainGift", g))
		giftViews[i] = domainGift.ConvertDomainGiftToProtoView(g)
	}

	return &giftv1.GetGiftsResponse{
		Gifts: giftViews,
		Pagination: &sharedv1.PageResponse{
			Page:       pagination.Page(),
			PageSize:   pagination.PageSize(),
			Total:      int32(len(domainGifts)),
			TotalPages: pagination.TotalPages(int32(len(domainGifts))),
		},
	}, nil
}

func (h *giftPublicHandler) WithdrawGift(ctx context.Context, req *giftv1.WithdrawGiftRequest) (*giftv1.WithdrawGiftResponse, error) {
	_, err := h.giftService.WithdrawGift(ctx, req.GetGiftId().Value)
	if err != nil {
		return nil, err
	}

	// For now, return success response
	// TODO: implement actual withdrawal logic based on method
	return &giftv1.WithdrawGiftResponse{
		Result: &giftv1.WithdrawGiftResponse_Success{
			Success: &sharedv1.SuccessResponse{
				Success: true,
				Message: "Withdrawal request received",
			},
		},
		WithdrawalId: "temp-withdrawal-id", // TODO: generate proper withdrawal ID
	}, nil
}
