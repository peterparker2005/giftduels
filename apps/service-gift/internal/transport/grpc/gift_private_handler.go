package grpc

import (
	"context"

	domainGift "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
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

func (h *giftPrivateHandler) PrivateGetGifts(ctx context.Context, req *giftv1.PrivateGetGiftsRequest) (*giftv1.PrivateGetGiftsResponse, error) {
	giftIDs := make([]string, len(req.GetGiftIds()))
	for i, giftID := range req.GetGiftIds() {
		giftIDs[i] = giftID.GetValue()
	}
	gifts, err := h.giftService.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		return nil, err
	}

	giftProtos := make([]*giftv1.Gift, len(gifts))
	for i, g := range gifts {
		giftProtos[i] = domainGift.DomainGiftToProto(g)
	}

	return &giftv1.PrivateGetGiftsResponse{
		Gifts: giftProtos,
	}, nil
}

func (h *giftPrivateHandler) GetUserGifts(ctx context.Context, req *giftv1.GetUserGiftsRequest) (*giftv1.GetUserGiftsResponse, error) {
	pagination := shared.NewPageRequest(req.GetPagination().GetPage(), req.GetPagination().GetPageSize())

	domainGifts, err := h.giftService.GetUserGifts(ctx, req.TelegramUserId.Value, pagination)
	if err != nil {
		return nil, err
	}

	// Convert domain gifts to protobuf
	giftViews := make([]*giftv1.Gift, len(domainGifts.Gifts))
	for i, g := range domainGifts.Gifts {
		giftViews[i] = domainGift.DomainGiftToProto(g)
	}

	return &giftv1.GetUserGiftsResponse{
		Gifts: giftViews,
		Pagination: &sharedv1.PageResponse{
			Page:       pagination.Page(),
			PageSize:   pagination.PageSize(),
			Total:      domainGifts.Total,
			TotalPages: pagination.TotalPages(domainGifts.Total),
		},
	}, nil
}

func (h *giftPrivateHandler) StakeGift(ctx context.Context, req *giftv1.StakeGiftRequest) (*giftv1.StakeGiftResponse, error) {
	g, err := h.giftService.StakeGift(ctx, req.GetGiftId().Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.StakeGiftResponse{
		Gift: domainGift.DomainGiftToProto(g),
	}, nil
}

func (h *giftPrivateHandler) TransferGiftToUser(ctx context.Context, req *giftv1.TransferGiftToUserRequest) (*giftv1.TransferGiftToUserResponse, error) {
	g, err := h.giftService.TransferGiftToUser(ctx, req.GetGiftId().Value, req.GetTelegramUserId().Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.TransferGiftToUserResponse{
		Gift: domainGift.DomainGiftToProto(g),
	}, nil
}

func (h *giftPrivateHandler) PrivateGetGift(ctx context.Context, req *giftv1.PrivateGetGiftRequest) (*giftv1.PrivateGetGiftResponse, error) {
	g, err := h.giftService.GetGiftByID(ctx, req.GetGiftId().Value)
	if err != nil {
		return nil, err
	}

	return &giftv1.PrivateGetGiftResponse{
		Gift: domainGift.DomainGiftToProto(g),
	}, nil
}
