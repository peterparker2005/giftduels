package grpchandlers

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/adapter/proto"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/service/user"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
)

var _ identityv1.IdentityPrivateServiceServer = (*IdentityPrivateHandler)(nil)

type IdentityPrivateHandler struct {
	identityv1.UnimplementedIdentityPrivateServiceServer

	userSvc *user.Service
	logger  *logger.Logger
}

func NewIdentityPrivateHandler(
	us *user.Service,
	lg *logger.Logger,
) identityv1.IdentityPrivateServiceServer {
	return &IdentityPrivateHandler{userSvc: us, logger: lg}
}

func (h *IdentityPrivateHandler) GetUserByID(
	ctx context.Context,
	req *identityv1.GetUserByIDRequest,
) (*identityv1.GetUserByIDResponse, error) {
	user, err := h.userSvc.GetUserByTelegramID(ctx, req.GetTelegramUserId().GetValue())
	if err != nil {
		return nil, err
	}

	return &identityv1.GetUserByIDResponse{User: proto.ToPBUser(user)}, nil
}

func (h *IdentityPrivateHandler) GetUsersByIDs(
	ctx context.Context,
	req *identityv1.GetUsersByIDsRequest,
) (*identityv1.GetUsersByIDsResponse, error) {
	telegramUserIDs := make([]int64, len(req.GetTelegramUserIds()))
	for i, telegramUserID := range req.GetTelegramUserIds() {
		telegramUserIDs[i] = telegramUserID.GetValue()
	}

	users, err := h.userSvc.GetUsersByTelegramIDs(ctx, telegramUserIDs)
	if err != nil {
		return nil, err
	}

	pbUsers := make([]*identityv1.User, len(users))
	for i, user := range users {
		pbUsers[i] = proto.ToPBUser(user)
	}

	return &identityv1.GetUsersByIDsResponse{Users: pbUsers}, nil
}
