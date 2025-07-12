package duel

import (
	"context"

	"github.com/google/uuid"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type DuelService struct {
	repo              duelDomain.Repository
	giftPrivateClient giftv1.GiftPrivateServiceClient
}

func NewDuelService(repo duelDomain.Repository, clients *clients.Clients) *DuelService {
	return &DuelService{repo: repo, giftPrivateClient: clients.Gift.Private}
}

type GetDuelListResponse struct {
	Duels []*duelDomain.Duel
	Total int64
}

func (s *DuelService) GetDuelList(ctx context.Context, pageRequest *shared.PageRequest) (*GetDuelListResponse, error) {
	duels, total, err := s.repo.GetDuelList(ctx, pageRequest)
	if err != nil {
		return nil, err
	}
	return &GetDuelListResponse{Duels: duels, Total: total}, nil
}

type CreateDuelParams struct {
	Params       duelDomain.DuelParams
	Participants []duelDomain.Participant
	Stakes       []duelDomain.Stake
}

func (s *DuelService) CreateDuel(ctx context.Context, telegramUserID int64, params CreateDuelParams) (duelDomain.ID, error) {
	gameID := uuid.New().String()
	for i, stake := range params.Stakes {
		gift, err := s.giftPrivateClient.StakeGift(ctx, &giftv1.StakeGiftRequest{
			GiftId:         &sharedv1.GiftId{Value: stake.GiftID},
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			GameMetadata: &giftv1.StakeGiftRequest_GameMetadata{
				GameMode: sharedv1.GameMode_GAME_MODE_DUEL,
				GameId:   gameID,
			},
		})
		if err != nil {
			return "", err
		}
		params.Stakes[i].StakeValue = gift.GetGift().GetPrice().GetValue()
		params.Stakes[i].TelegramUserID = duelDomain.TelegramUserID(telegramUserID)
	}
	duelID, err := duelDomain.NewID(gameID)
	if err != nil {
		return "", err
	}
	duelID, err = s.repo.CreateDuel(ctx, duelDomain.CreateDuelParams{
		DuelID:       duelID,
		Params:       params.Params,
		Participants: params.Participants,
		Stakes:       params.Stakes,
	})
	if err != nil {
		return "", err
	}
	return duelID, nil
}
