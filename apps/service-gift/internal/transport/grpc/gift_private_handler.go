package grpc

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/query"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/saga"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type giftPrivateHandler struct {
	giftv1.GiftPrivateServiceServer

	withdrawalSaga      *saga.WithdrawalSaga
	giftStakeCommand    *command.GiftStakeCommand
	giftWithdrawCommand *command.GiftWithdrawCommand
	giftReadService     *query.GiftReadService
	userGiftsService    *query.UserGiftsService
}

func NewGiftPrivateHandler(withdrawalSaga *saga.WithdrawalSaga) giftv1.GiftPrivateServiceServer {
	return &giftPrivateHandler{
		withdrawalSaga: withdrawalSaga,
	}
}

func (h *giftPrivateHandler) PrivateGetGifts(
	ctx context.Context,
	req *giftv1.PrivateGetGiftsRequest,
) (*giftv1.PrivateGetGiftsResponse, error) {
	giftIDs := make([]string, len(req.GetGiftIds()))
	for i, giftID := range req.GetGiftIds() {
		giftIDs[i] = giftID.GetValue()
	}
	gifts, err := h.giftReadService.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		return nil, err
	}

	giftProtos := make([]*giftv1.Gift, len(gifts))
	for i, g := range gifts {
		giftProtos[i] = DomainGiftToProto(g)
	}

	return &giftv1.PrivateGetGiftsResponse{
		Gifts: giftProtos,
	}, nil
}

func (h *giftPrivateHandler) GetUserGifts(
	ctx context.Context,
	req *giftv1.GetUserGiftsRequest,
) (*giftv1.GetUserGiftsResponse, error) {
	pagination := shared.NewPageRequest(
		req.GetPagination().GetPage(),
		req.GetPagination().GetPageSize(),
	)

	domainGifts, err := h.userGiftsService.GetUserGifts(
		ctx,
		req.GetTelegramUserId().GetValue(),
		pagination,
	)
	if err != nil {
		return nil, err
	}

	// Convert domain gifts to protobuf
	giftViews := make([]*giftv1.Gift, len(domainGifts.Gifts))
	for i, g := range domainGifts.Gifts {
		giftViews[i] = DomainGiftToProto(g)
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

func (h *giftPrivateHandler) StakeGift(
	ctx context.Context,
	req *giftv1.StakeGiftRequest,
) (*giftv1.StakeGiftResponse, error) {
	g, err := h.giftStakeCommand.StakeGift(ctx, command.StakeGiftParams{
		GiftID:         req.GetGiftId().GetValue(),
		TelegramUserID: req.GetTelegramUserId().GetValue(),
		GameMetadata:   req.GetGameMetadata(),
	})
	if err != nil {
		return nil, err
	}

	return &giftv1.StakeGiftResponse{
		Gift: DomainGiftToProto(g),
	}, nil
}

// func (h *giftPrivateHandler) TransferGiftToUser(ctx context.Context, req *giftv1.TransferGiftToUserRequest) (*giftv1.TransferGiftToUserResponse, error) {
// 	g, err := h.giftService.TransferGiftToUser(ctx, req.GetGiftId().Value, req.GetTelegramUserId().Value)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &giftv1.TransferGiftToUserResponse{
// 		Gift: DomainGiftToProto(g),
// 	}, nil
// }

func (h *giftPrivateHandler) PrivateGetGift(
	ctx context.Context,
	req *giftv1.PrivateGetGiftRequest,
) (*giftv1.PrivateGetGiftResponse, error) {
	g, err := h.giftReadService.GetGiftByID(ctx, req.GetGiftId().GetValue())
	if err != nil {
		return nil, err
	}

	return &giftv1.PrivateGetGiftResponse{
		Gift: DomainGiftToProto(g),
	}, nil
}
