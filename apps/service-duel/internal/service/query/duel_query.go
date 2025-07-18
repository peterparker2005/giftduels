package query

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
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
				s.log.Error("failed to get gift", zap.Error(err))
				continue
			}
			for i := range duel.Stakes {
				price, err := tonamount.NewTonAmountFromString(resp.Gifts[i].GetPrice().GetValue())
				if err != nil {
					s.log.Error("failed to parse price", zap.Error(err))
					continue
				}
				gift, err := dueldomain.NewStakedGiftBuilder().
					WithID(resp.GetGifts()[i].GetGiftId().GetValue()).
					WithTitle(resp.GetGifts()[i].GetTitle()).
					WithSlug(resp.GetGifts()[i].GetSlug()).
					WithPrice(price).
					Build()
				if err != nil {
					s.log.Error("failed to build staked gift", zap.Error(err))
					continue
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
			continue
		}
		for i, user := range userResp.GetUsers() {
			duel.Participants[i].PhotoURL = user.GetPhotoUrl()
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
		resp, err := s.giftPrivateClient.PrivateGetGift(ctx, &giftv1.PrivateGetGiftRequest{
			GiftId: &sharedv1.GiftId{Value: stakeIDs[0].Value},
		})
		if err != nil {
			s.log.Error("failed to get gift", zap.Error(err))
			return nil, err
		}
		for i := range duel.Stakes {
			price, err := tonamount.NewTonAmountFromString(resp.GetGift().GetPrice().GetValue())
			if err != nil {
				s.log.Error("failed to parse price", zap.Error(err))
				return nil, err
			}
			gift, err := dueldomain.NewStakedGiftBuilder().
				WithID(resp.GetGift().GetGiftId().GetValue()).
				WithTitle(resp.GetGift().GetTitle()).
				WithSlug(resp.GetGift().GetSlug()).
				WithPrice(price).
				Build()
			if err != nil {
				s.log.Error("failed to build staked gift", zap.Error(err))
				return nil, err
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
			return nil, userErr
		}
		for i, user := range userResp.GetUsers() {
			duel.Participants[i].PhotoURL = user.GetPhotoUrl()
		}
	}

	return duel, nil
}
