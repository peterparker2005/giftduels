package grpc

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
)

type giftPrivateHandler struct {
	giftv1.GiftPrivateServiceServer

	giftService *gift.Service
}

func NewGiftPrivateHandler(giftService *gift.Service) giftv1.GiftPrivateServiceServer {
	return &giftPrivateHandler{
		giftService: giftService,
	}
}

func (h *giftPrivateHandler) GetUserGifts(ctx context.Context, req *giftv1.GetUserGiftsRequest) (*giftv1.GetUserGiftsResponse, error) {
	limit := int32(10) // default limit
	offset := int32(0)

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			limit = int32(req.Pagination.PageSize)
		}
		if req.Pagination.Page > 0 {
			offset = int32((req.Pagination.Page - 1) * req.Pagination.PageSize)
		}
	}

	domainGifts, err := h.giftService.GetUserGifts(ctx, req.TelegramUserId.Value, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert domain gifts to protobuf
	giftViews := make([]*giftv1.Gift, len(domainGifts))
	for i, domainGift := range domainGifts {
		giftViews[i] = ConvertDomainGiftToProto(domainGift)
	}

	return &giftv1.GetUserGiftsResponse{
		Gifts: giftViews,
		Pagination: &sharedv1.PageResponse{
			Page:     req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
			// TODO: implement total count
		},
	}, nil
}

func (h *giftPrivateHandler) StakeGift(ctx context.Context, req *giftv1.StakeGiftRequest) (*giftv1.StakeGiftResponse, error) {
	domainGift, err := h.giftService.StakeGift(ctx, req.GiftId.Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.StakeGiftResponse{
		Gift: ConvertDomainGiftToProto(domainGift),
	}, nil
}

func (h *giftPrivateHandler) TransferGiftToUser(ctx context.Context, req *giftv1.TransferGiftToUserRequest) (*giftv1.TransferGiftToUserResponse, error) {
	domainGift, err := h.giftService.TransferGiftToUser(ctx, req.GiftId.Value, req.TelegramUserId.Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.TransferGiftToUserResponse{
		Gift: ConvertDomainGiftToProto(domainGift),
	}, nil
}

func (h *giftPrivateHandler) PrivateGetGift(ctx context.Context, req *giftv1.PrivateGetGiftRequest) (*giftv1.PrivateGetGiftResponse, error) {
	domainGift, err := h.giftService.GetGiftByID(ctx, req.GiftId.Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.PrivateGetGiftResponse{
		Gift: ConvertDomainGiftToProto(domainGift),
	}, nil
}
