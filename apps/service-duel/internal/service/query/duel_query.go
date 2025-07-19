package query

import (
	"context"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/authctx"
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
	giftPrivateClient     giftv1.GiftPrivateServiceClient
	identityPrivateClient identityv1.IdentityPrivateServiceClient
}

type GetDuelListResponse struct {
	Duels []*dueldomain.Duel
	Total int64
}

func NewDuelQueryService(
	log *logger.Logger,
	repo dueldomain.Repository,
	giftPrivateClient giftv1.GiftPrivateServiceClient,
	identityPrivateClient identityv1.IdentityPrivateServiceClient,
) *DuelQueryService {
	return &DuelQueryService{
		log:                   log,
		repo:                  repo,
		giftPrivateClient:     giftPrivateClient,
		identityPrivateClient: identityPrivateClient,
	}
}

func (s *DuelQueryService) GetDuelList(
	ctx context.Context,
	pageRequest *shared.PageRequest,
	filter *dueldomain.Filter,
) (*GetDuelListResponse, error) {
	userID, err := authctx.TelegramUserID(ctx)
	if err != nil {
		s.log.Error("failed to get telegram user id", zap.Error(err))
		return nil, err
	}
	domainUserID, err := dueldomain.NewTelegramUserID(userID)
	if err != nil {
		s.log.Error("failed to create telegram user id", zap.Error(err))
		return nil, err
	}

	duels, total, err := s.repo.GetDuelList(ctx, pageRequest, filter, &domainUserID)
	if err != nil {
		s.log.Error("failed to get duel list", zap.Error(err))
		if pg.IsNotFound(err) {
			return nil, ErrDuelNotFound
		}
		return nil, ErrDatabase
	}

	if err = s.enrichStakes(ctx, duels); err != nil {
		return nil, err
	}
	if err = s.enrichParticipants(ctx, duels); err != nil {
		return nil, err
	}

	return &GetDuelListResponse{Duels: duels, Total: total}, nil
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

	// обогащаем по тем же helper-методам, но для одного элемента
	if err = s.enrichStakes(ctx, []*dueldomain.Duel{duel}); err != nil {
		return nil, err
	}
	if err = s.enrichParticipants(ctx, []*dueldomain.Duel{duel}); err != nil {
		return nil, err
	}

	return duel, nil
}

// enrichStakes подтягивает из gift-сервиса полные объекты Gift для всех ставок.
func (s *DuelQueryService) enrichStakes(
	ctx context.Context,
	duels []*dueldomain.Duel,
) error {
	for _, duel := range duels {
		if len(duel.Stakes) == 0 {
			continue
		}
		// собираем все GiftId
		ids := make([]*sharedv1.GiftId, len(duel.Stakes))
		for i, st := range duel.Stakes {
			ids[i] = &sharedv1.GiftId{Value: st.Gift.ID}
		}
		resp, err := s.giftPrivateClient.PrivateGetGifts(ctx, &giftv1.PrivateGetGiftsRequest{
			GiftIds: ids,
		})
		if err != nil {
			s.log.Error("failed to get gifts", zap.Error(err))
			return ErrGetGifts
		}
		// map[id]→giftProto
		giftMap := make(map[string]*giftv1.Gift, len(resp.GetGifts()))
		for _, g := range resp.GetGifts() {
			giftMap[g.GetGiftId().GetValue()] = g
		}
		// обновляем каждый stake
		for i := range duel.Stakes {
			protoGift, ok := giftMap[duel.Stakes[i].Gift.ID]
			if !ok {
				s.log.Error("gift not found in response",
					zap.String("giftID", duel.Stakes[i].Gift.ID))
				return ErrGiftNotFoundInResponse
			}
			price, perr := tonamount.NewTonAmountFromString(protoGift.GetPrice().GetValue())
			if perr != nil {
				s.log.Error("parse price failed", zap.Error(perr))
				return ErrParseGiftPrice
			}
			gdom, berr := dueldomain.NewStakedGiftBuilder().
				WithID(protoGift.GetGiftId().GetValue()).
				WithTitle(protoGift.GetTitle()).
				WithSlug(protoGift.GetSlug()).
				WithPrice(price).
				Build()
			if berr != nil {
				s.log.Error("build staked gift failed", zap.Error(berr))
				return ErrBuildStakedGift
			}
			duel.Stakes[i].Gift = gdom
		}
	}
	return nil
}

// enrichParticipants подтягивает photo_url для всех участников из identity-сервиса.
func (s *DuelQueryService) enrichParticipants(
	ctx context.Context,
	duels []*dueldomain.Duel,
) error {
	for _, duel := range duels {
		if len(duel.Participants) == 0 {
			continue
		}
		ids := make([]*sharedv1.TelegramUserId, len(duel.Participants))
		for i, p := range duel.Participants {
			ids[i] = &sharedv1.TelegramUserId{Value: p.TelegramUserID.Int64()}
		}
		ur, err := s.identityPrivateClient.GetUsersByIDs(ctx, &identityv1.GetUsersByIDsRequest{
			TelegramUserIds: ids,
		})
		if err != nil {
			s.log.Error("failed to get users", zap.Error(err))
			return ErrGetUsers
		}
		photoMap := make(map[int64]string, len(ur.GetUsers()))
		for _, u := range ur.GetUsers() {
			photoMap[u.GetTelegramId().GetValue()] = u.GetPhotoUrl()
		}
		for i := range duel.Participants {
			id := duel.Participants[i].TelegramUserID.Int64()
			if url, ok := photoMap[id]; ok {
				duel.Participants[i].PhotoURL = url
			} else {
				s.log.Warn("photo missing", zap.Int64("userID", id))
			}
		}
	}
	return nil
}

func (s *DuelQueryService) FindDuelByGiftID(
	ctx context.Context,
	giftID string,
) (dueldomain.ID, error) {
	return s.repo.FindDuelByGiftID(ctx, giftID)
}
