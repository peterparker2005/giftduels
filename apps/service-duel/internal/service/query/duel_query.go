package query

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/errors/pkg/errors"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

type DuelQueryService struct {
	log                   *logger.Logger
	repo                  dueldomain.Repository
	txManager             pg.TxManager
	giftPrivateClient     giftv1.GiftPrivateServiceClient
	identityPrivateClient identityv1.IdentityPrivateServiceClient
}

func NewDuelQueryService(
	repo dueldomain.Repository,
	clients *clients.Clients,
	txManager pg.TxManager,
	log *logger.Logger,
) *DuelQueryService {
	return &DuelQueryService{
		repo:                  repo,
		txManager:             txManager,
		log:                   log,
		giftPrivateClient:     clients.Gift.Private,
		identityPrivateClient: clients.Identity.Private,
	}
}

type GetDuelListResponse struct {
	Duels []*dueldomain.Duel
	Total int64
}

func (s *DuelQueryService) GetDuelList(
	ctx context.Context,
	pageRequest *shared.PageRequest,
) (*GetDuelListResponse, error) {
	duels, total, err := s.repo.GetDuelList(ctx, pageRequest)
	if err != nil {
		s.log.Error("failed to get duel list", zap.Error(err))
		return nil, err
	}
	for _, duel := range duels {
		if len(duel.Stakes) > 0 {
			stakeIDs := make([]*sharedv1.GiftId, len(duel.Stakes))
			for i := range duel.Stakes {
				stakeIDs[i] = &sharedv1.GiftId{Value: duel.Stakes[i].Gift.ID}
			}
			resp, err := s.giftPrivateClient.PrivateGetGifts(ctx, &giftv1.PrivateGetGiftsRequest{
				GiftIds: stakeIDs,
			})
			if err != nil {
				s.log.Error("failed to get gifts", zap.Error(err))
				continue
			}

			// Create a map for quick lookup
			giftMap := make(map[string]*giftv1.Gift)
			for _, gift := range resp.GetGifts() {
				giftMap[gift.GetGiftId().GetValue()] = gift
			}

			for i := range duel.Stakes {
				giftProto, exists := giftMap[duel.Stakes[i].Gift.ID]
				if !exists {
					s.log.Error("gift not found in response", zap.String("giftID", duel.Stakes[i].Gift.ID))
					return nil, errors.NewInternalError("gift not found in response")
				}

				price, err := tonamount.NewTonAmountFromString(giftProto.GetPrice().GetValue())
				if err != nil {
					s.log.Error("failed to parse price", zap.Error(err))
					return nil, errors.NewInternalError("failed to parse price")
				}
				gift, err := dueldomain.NewStakedGiftBuilder().
					WithID(giftProto.GetGiftId().GetValue()).
					WithTitle(giftProto.GetTitle()).
					WithSlug(giftProto.GetSlug()).
					WithPrice(price).
					Build()
				if err != nil {
					s.log.Error("failed to build staked gift", zap.Error(err))
					return nil, errors.NewInternalError("failed to build staked gift")
				}
				duel.Stakes[i].Gift = gift
			}
		}
		telegramUserIDs := make([]*sharedv1.TelegramUserId, len(duel.Participants))
		for i := range duel.Participants {
			telegramUserIDs[i] = &sharedv1.TelegramUserId{
				Value: duel.Participants[i].TelegramUserID.Int64(),
			}
		}

		userResp, userErr := s.identityPrivateClient.GetUsersByIDs(
			ctx,
			&identityv1.GetUsersByIDsRequest{
				TelegramUserIds: telegramUserIDs,
			},
		)
		if userErr != nil {
			s.log.Error("failed to get user", zap.Error(userErr))
			return nil, errors.NewInternalError("failed to get user")
		}
		userPhoto := make(map[int64]string, len(userResp.GetUsers()))
		for _, u := range userResp.GetUsers() {
			id := u.GetTelegramId().GetValue()
			userPhoto[id] = u.GetPhotoUrl()
		}

		for i := range duel.Participants {
			id := duel.Participants[i].TelegramUserID.Int64()
			if photo, ok := userPhoto[id]; ok {
				duel.Participants[i].PhotoURL = photo
			} else {
				// на случай, если вдруг пользователь не вернулся в ответе
				s.log.Warn("photo for user not found", zap.Int64("userID", id))
			}
		}
	}
	return &GetDuelListResponse{Duels: duels, Total: total}, nil
}

func (s *DuelQueryService) FindDuelByGiftID(
	ctx context.Context,
	giftID string,
) (dueldomain.ID, error) {
	return s.repo.FindDuelByGiftID(ctx, giftID)
}

func (s *DuelQueryService) GetDuelByID(
	ctx context.Context,
	duelID dueldomain.ID,
) (*dueldomain.Duel, error) {
	duel, err := s.repo.GetDuelByID(ctx, duelID)
	if err != nil {
		s.log.Error("failed to get duel by id", zap.Error(err))
		return nil, err
	}

	// Load gift data for stakes
	if len(duel.Stakes) > 0 {
		stakeIDs := make([]*sharedv1.GiftId, len(duel.Stakes))
		for i := range duel.Stakes {
			stakeIDs[i] = &sharedv1.GiftId{Value: duel.Stakes[i].Gift.ID}
		}
		resp, err := s.giftPrivateClient.PrivateGetGifts(ctx, &giftv1.PrivateGetGiftsRequest{
			GiftIds: stakeIDs,
		})
		if err != nil {
			s.log.Error("failed to get gifts", zap.Error(err))
			return nil, errors.NewInternalError("failed to get gifts")
		}

		// Create a map for quick lookup
		giftMap := make(map[string]*giftv1.Gift)
		for _, gift := range resp.GetGifts() {
			giftMap[gift.GetGiftId().GetValue()] = gift
		}

		for i := range duel.Stakes {
			giftProto, exists := giftMap[duel.Stakes[i].Gift.ID]
			if !exists {
				s.log.Error("gift not found in response", zap.String("giftID", duel.Stakes[i].Gift.ID))
				return nil, errors.NewInternalError("gift not found in response")
			}

			price, err := tonamount.NewTonAmountFromString(giftProto.GetPrice().GetValue())
			if err != nil {
				s.log.Error("failed to parse price", zap.Error(err))
				return nil, errors.NewInternalError("failed to parse price")
			}
			gift, err := dueldomain.NewStakedGiftBuilder().
				WithID(giftProto.GetGiftId().GetValue()).
				WithTitle(giftProto.GetTitle()).
				WithSlug(giftProto.GetSlug()).
				WithPrice(price).
				Build()
			if err != nil {
				s.log.Error("failed to build staked gift", zap.Error(err))
				return nil, errors.NewInternalError("failed to build staked gift")
			}
			duel.Stakes[i].Gift = gift
		}
	}

	// Load participant data
	if len(duel.Participants) > 0 {
		telegramUserIDs := make([]*sharedv1.TelegramUserId, len(duel.Participants))
		for i := range duel.Participants {
			telegramUserIDs[i] = &sharedv1.TelegramUserId{
				Value: duel.Participants[i].TelegramUserID.Int64(),
			}
		}

		userResp, userErr := s.identityPrivateClient.GetUsersByIDs(
			ctx,
			&identityv1.GetUsersByIDsRequest{
				TelegramUserIds: telegramUserIDs,
			},
		)
		if userErr != nil {
			s.log.Error("failed to get user", zap.Error(userErr))
			return nil, errors.NewInternalError("failed to get user")
		}
		userPhoto := make(map[int64]string, len(userResp.GetUsers()))
		for _, u := range userResp.GetUsers() {
			id := u.GetTelegramId().GetValue()
			userPhoto[id] = u.GetPhotoUrl()
		}

		for i := range duel.Participants {
			id := duel.Participants[i].TelegramUserID.Int64()
			if photo, ok := userPhoto[id]; ok {
				duel.Participants[i].PhotoURL = photo
			} else {
				// на случай, если вдруг пользователь не вернулся в ответе
				s.log.Warn("photo for user not found", zap.Int64("userID", id))
			}
		}
	}

	return duel, nil
}
