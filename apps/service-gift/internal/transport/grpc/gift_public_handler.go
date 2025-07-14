package grpc

import (
	"context"

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

// NewGiftPublicHandler создает новый GRPC handler.
func NewGiftPublicHandler(
	giftService *gift.Service,
	logger *logger.Logger,
) giftv1.GiftPublicServiceServer {
	return &giftPublicHandler{
		giftService: giftService,
		logger:      logger,
	}
}

func (h *giftPublicHandler) GetGift(
	ctx context.Context,
	req *giftv1.GetGiftRequest,
) (*giftv1.GetGiftResponse, error) {
	g, err := h.giftService.GetGiftByID(ctx, req.GetGiftId().GetValue())
	if err != nil {
		return nil, err
	}

	return &giftv1.GetGiftResponse{
		Gift: DomainGiftToProtoView(g),
	}, nil
}

func (h *giftPublicHandler) GetGifts(
	ctx context.Context,
	req *giftv1.GetGiftsRequest,
) (*giftv1.GetGiftsResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	log := h.logger.With(zap.Int64("telegramUserID", telegramUserID))

	pagination := shared.NewPageRequest(
		req.GetPagination().GetPage(),
		req.GetPagination().GetPageSize(),
	)
	domainGifts, err := h.giftService.GetUserActiveGifts(ctx, telegramUserID, pagination)
	if err != nil {
		log.Error("Failed to get user active gifts", zap.Error(err))
		return nil, err
	}

	giftViews := make([]*giftv1.GiftView, len(domainGifts.Gifts))
	for i, g := range domainGifts.Gifts {
		giftViews[i] = DomainGiftToProtoView(g)
	}

	return &giftv1.GetGiftsResponse{
		Gifts: giftViews,
		TotalValue: &sharedv1.TonAmount{
			Value: domainGifts.TotalValue.String(),
		},
		Pagination: &sharedv1.PageResponse{
			Page:       pagination.Page(),
			PageSize:   pagination.PageSize(),
			Total:      domainGifts.Total,
			TotalPages: pagination.TotalPages(domainGifts.Total),
		},
	}, nil
}

func (h *giftPublicHandler) ExecuteWithdraw(
	ctx context.Context,
	req *giftv1.ExecuteWithdrawRequest,
) (*giftv1.ExecuteWithdrawResponse, error) {
	telegramUserID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(req.GetGiftIds()))
	for i, id := range req.GetGiftIds() {
		ids[i] = id.GetValue()
	}

	result, err := h.giftService.ExecuteWithdraw(
		ctx,
		telegramUserID,
		ids,
		req.GetCommissionCurrency(),
	)
	if err != nil {
		return nil, err
	}

	if result.IsStarsCommission {
		// Возвращаем Stars invoice URL
		return &giftv1.ExecuteWithdrawResponse{
			Response: &giftv1.ExecuteWithdrawResponse_StarsInvoiceUrl{
				StarsInvoiceUrl: result.StarsInvoiceURL,
			},
		}, nil
	}
	// Возвращаем успешный ответ для TON
	return &giftv1.ExecuteWithdrawResponse{
		Response: &giftv1.ExecuteWithdrawResponse_TonSuccess{
			TonSuccess: &sharedv1.SuccessResponse{
				Success: true,
				Message: "Withdrawal request received",
			},
		},
	}, nil
}
