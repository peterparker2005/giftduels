package grpc

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	"github.com/peterparker2005/giftduels/apps/service-identity/pkg/grpc"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type giftPublicHandler struct {
	// встраиваем заглушку из protobuf, чтобы не ломать backward-совместимость при добавлении новых методов
	giftv1.GiftPublicServiceServer

	// зависимость от сервисного слоя
	giftService *gift.Service
}

// NewGiftPublicHandler создает новый GRPC handler
func NewGiftPublicHandler(giftService *gift.Service) giftv1.GiftPublicServiceServer {
	return &giftPublicHandler{
		giftService: giftService,
	}
}

func (h *giftPublicHandler) GetGift(ctx context.Context, req *giftv1.GetGiftRequest) (*giftv1.GetGiftResponse, error) {
	domainGift, err := h.giftService.GetGiftByID(ctx, req.GetGiftId().Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.GetGiftResponse{
		Gift: ConvertDomainGiftToProtoView(domainGift),
	}, nil
}

func (h *giftPublicHandler) GetGifts(ctx context.Context, req *giftv1.GetGiftsRequest) (*giftv1.GetGiftsResponse, error) {
	telegramUserID, err := grpc.TelegramUserID(ctx)
	if err != nil {
		return nil, err
	}

	pagination := shared.NewPageRequest(req.Pagination.Page, req.Pagination.PageSize)
	domainGifts, err := h.giftService.GetUserGifts(ctx, telegramUserID, pagination.PageSize(), pagination.Offset())
	if err != nil {
		return nil, err
	}

	giftViews := make([]*giftv1.GiftView, len(domainGifts))
	for i, domainGift := range domainGifts {
		giftViews[i] = ConvertDomainGiftToProtoView(domainGift)
	}

	return &giftv1.GetGiftsResponse{
		Gifts: giftViews,
		Pagination: &sharedv1.PageResponse{
			Page:       req.Pagination.Page,
			PageSize:   req.Pagination.PageSize,
			Total:      int32(len(domainGifts)), // TODO: implement proper total count
			TotalPages: 1,                       // TODO: implement proper total pages
		},
	}, nil
}

func (h *giftPublicHandler) WithdrawGift(ctx context.Context, req *giftv1.WithdrawGiftRequest) (*giftv1.WithdrawGiftResponse, error) {
	_, err := h.giftService.WithdrawGift(ctx, req.GiftId.Value)
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

func (h *giftPublicHandler) GetWithdrawOptions(ctx context.Context, req *giftv1.GetWithdrawOptionsRequest) (*giftv1.GetWithdrawOptionsResponse, error) {
	// Return default withdrawal options
	return &giftv1.GetWithdrawOptionsResponse{
		Options: []*giftv1.WithdrawOption{
			{
				Method:      giftv1.WithdrawMethod_WITHDRAW_METHOD_STARS_PAYMENT,
				StarsCost:   &sharedv1.StarsAmount{Value: 100},
				Description: "Withdraw for Stars",
				IsAvailable: true,
			},
		},
	}, nil
}
