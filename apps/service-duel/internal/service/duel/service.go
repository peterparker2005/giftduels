package duel

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	identityv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/identity/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

type Service struct {
	log                   *logger.Logger
	publisher             message.Publisher
	repo                  dueldomain.Repository
	txManager             pg.TxManager
	giftPrivateClient     giftv1.GiftPrivateServiceClient
	telegramPrivateClient telegrambotv1.TelegramBotPrivateServiceClient
	identityPrivateClient identityv1.IdentityPrivateServiceClient
}

func NewDuelService(
	repo dueldomain.Repository,
	clients *clients.Clients,
	txManager pg.TxManager,
	log *logger.Logger,
	publisher message.Publisher,
) *Service {
	return &Service{
		repo:                  repo,
		txManager:             txManager,
		log:                   log,
		giftPrivateClient:     clients.Gift.Private,
		identityPrivateClient: clients.Identity.Private,
		publisher:             publisher,
	}
}

type GetDuelListResponse struct {
	Duels []*dueldomain.Duel
	Total int64
}

func (s *Service) GetDuelList(
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
				duel.Stakes[i].Gift = dueldomain.NewStakedGift(
					resp.GetGifts()[i].GetGiftId().GetValue(),
					resp.GetGifts()[i].GetTitle(),
					resp.GetGifts()[i].GetSlug(),
					price,
				)
			}
		}
		userResp, err := s.identityPrivateClient.GetUsersByIDs(ctx, &identityv1.GetUsersByIDsRequest{
			TelegramUserIds: []*sharedv1.TelegramUserId{
				{Value: duel.Participants[0].TelegramUserID.Int64()},
			},
		})
		for i, user := range userResp.GetUsers() {
			duel.Participants[i].PhotoURL = user.GetPhotoUrl()
		}
		if err != nil {
			s.log.Error("failed to get user", zap.Error(err))
			continue
		}
	}
	return &GetDuelListResponse{Duels: duels, Total: total}, nil
}

type CreateDuelParams struct {
	Params dueldomain.Params
	Stakes []dueldomain.Stake
}

func (s *Service) CreateDuel(
	ctx context.Context,
	telegramUserID int64,
	params CreateDuelParams,
) (dueldomain.ID, error) {
	// 1. Начинаем транзакцию
	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		// rollback, если в err что-то попало
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				s.log.Error("failed to rollback tx", zap.Error(err))
			}
		}
	}()

	repo := s.repo.WithTx(tx)

	duel := dueldomain.NewDuel(params.Params)
	creator := dueldomain.NewParticipant(dueldomain.TelegramUserID(telegramUserID), "", true)

	// 3. Доменный метод Join (валидация и добавление участника)
	if err = duel.Join(creator); err != nil {
		return "", err
	}

	// Track staked gifts for rollback in case of error
	stakedGiftIDs := make([]string, 0, len(params.Stakes))

	// 4. Для каждой ставки вызываем внешний gRPC (giftPrivateClient) и
	//    заполняем stakeValue из ответа
	for i, stake := range params.Stakes {
		s.log.Info("staking gift for duel",
			zap.String("giftID", stake.Gift.ID),
			zap.Int64("telegramUserID", telegramUserID))

		resp, stakeErr := s.giftPrivateClient.StakeGift(ctx, &giftv1.StakeGiftRequest{
			GiftId:         &sharedv1.GiftId{Value: stake.Gift.ID},
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			// …метаданные игры…
		})
		if stakeErr != nil {
			s.log.Error("failed to stake gift",
				zap.String("giftID", stake.Gift.ID),
				zap.Int64("telegramUserID", telegramUserID),
				zap.Error(stakeErr))

			return "", stakeErr
		}

		// Track successfully staked gift
		stakedGiftIDs = append(stakedGiftIDs, stake.Gift.ID)

		price, priceErr := tonamount.NewTonAmountFromString(resp.GetGift().GetPrice().GetValue())
		if priceErr != nil {
			return "", priceErr
		}
		// 4.1 обновляем stake в срезе
		params.Stakes[i].StakeValue = price
		params.Stakes[i].TelegramUserID = dueldomain.TelegramUserID(telegramUserID)
		// 4.2 доменный метод PlaceStake (валидация и обновление TotalStakeValue)
		if stakeErr = duel.PlaceStake(params.Stakes[i]); stakeErr != nil {
			return "", stakeErr
		}
	}

	// 5. Сохраняем Duel
	duelID, err := repo.CreateDuel(ctx, dueldomain.CreateDuelParams{
		DuelID:          duel.ID,
		TotalStakeValue: duel.TotalStakeValue,
		Params:          duel.Params,
	})
	if err != nil {
		s.log.Error("failed to create duel", zap.Error(err))
		return "", err
	}

	// 6. Сохраняем участника
	if err = repo.CreateParticipant(ctx, duelID, creator); err != nil {
		s.log.Error("failed to create participant", zap.Error(err))
		return "", err
	}

	// 7. Сохраняем все Stakes
	for _, stake := range params.Stakes {
		if err = repo.CreateStake(ctx, duelID, stake); err != nil {
			s.log.Error("failed to create stake", zap.Error(err))
			return "", err
		}
	}

	// 8. Коммит транзакции
	if err = tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", zap.Error(err))
		// Return staked gifts to owned status
		if publishErr := s.returnStakedGifts(stakedGiftIDs); publishErr != nil {
			s.log.Error("failed to return staked gifts", zap.Error(publishErr))
		}
		return "", err
	}

	return duelID, nil
}

// returnStakedGifts returns gifts from in_game status back to owned status.
func (s *Service) returnStakedGifts(giftIDs []string) error {
	for _, giftID := range giftIDs {
		s.log.Info("returning gift from game due to error", zap.String("giftID", giftID))
		// Note: We're not handling errors here as this is cleanup code
		// The main error is already being returned
		id := uuid.New().String()
		msg := message.NewMessage(id, []byte(giftID))
		err := s.publisher.Publish("gift.returned", msg)
		if err != nil {
			s.log.Error("failed to return gift from game during cleanup",
				zap.String("giftID", giftID),
				zap.Error(err))
		}
	}
	return nil
}

func (s *Service) rollDice(
	ctx context.Context,
	duelID *dueldomain.ID,
	telegramUserID int64,
) (int32, error) {
	resp, err := s.telegramPrivateClient.RollDice(ctx, &telegrambotv1.RollDiceRequest{
		RollerTelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
		Metadata: &telegrambotv1.RollDiceRequest_Metadata{
			Game: &telegrambotv1.RollDiceRequest_Metadata_Duel_{
				Duel: &telegrambotv1.RollDiceRequest_Metadata_Duel{
					DuelId: &sharedv1.DuelId{Value: duelID.String()},
				},
			},
		},
	})
	if err != nil {
		return 0, err
	}
	return resp.GetValue(), nil
}

func (s *Service) HandleAutoRoll(ctx context.Context, duelID dueldomain.ID) error {
	tx, err := s.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	repo := s.repo.WithTx(tx)

	d, err := repo.GetDuelByID(ctx, duelID)
	if err != nil {
		return err
	}

	// 1) в домене — проверяем, что раунд ждёт бросков
	last := d.Rounds[len(d.Rounds)-1]
	// для каждого игрока, у кого нет броска — делаем авто-бросок
	for _, pl := range last.Participants {
		already := false
		for _, rl := range last.Rolls {
			if rl.TelegramUserID == pl {
				already = true
				break
			}
		}
		if !already {
			// генерируем случайное от 1 до 6
			val, err := s.rollDice(ctx, &duelID, pl.Int64())
			if err != nil {
				return err
			}
			roll := dueldomain.NewRoll(pl, val, time.Now(), true)
			if err := d.AddRollToCurrentRound(roll); err != nil {
				return err
			}
			if err := repo.CreateRoll(ctx, duelID, roll); err != nil {
				return err
			}
		}
	}

	// 2) оцениваем раунд и либо завершаем дуэль, либо стартуем новый
	winners, finished := d.EvaluateCurrentRound()
	if !finished {
		// возможно пришёл ещё чей-то manual-roll раньше дедлайна
		// обновляем дедлайн, если надо
		next := time.Now().Add(d.TimeoutForRound())
		d.NextRollDeadline = &next
		if err := repo.UpdateNextRollDeadline(ctx, duelID, next); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	if len(winners) > 1 {
		// ничья — новый раунд среди tied
		if err := s.startNewRound(ctx, tx, d); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	// один победитель — завершаем дуэль
	if err := d.Complete(winners[0]); err != nil {
		return err
	}
	if err := repo.UpdateDuelStatus(ctx, duelID, d.Status, d.WinnerID, d.CompletedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) startNewRound(ctx context.Context, tx pgx.Tx, d *dueldomain.Duel) error {
	repo := s.repo.WithTx(tx)
	participants := make([]dueldomain.TelegramUserID, len(d.Participants))
	for i, p := range d.Participants {
		participants[i] = p.TelegramUserID
	}
	round := dueldomain.NewRound(len(d.Rounds)+1, participants)
	d.StartRound(participants)
	if err := repo.CreateRound(ctx, d.ID, round); err != nil {
		return err
	}
	return nil
}

func (s *Service) FindDuelByGiftID(ctx context.Context, giftID string) (dueldomain.ID, error) {
	return s.repo.FindDuelByGiftID(ctx, giftID)
}

func (s *Service) GetDuelByID(ctx context.Context, duelID dueldomain.ID) (*dueldomain.Duel, error) {
	duel, err := s.repo.GetDuelByID(ctx, duelID)
	if err != nil {
		s.log.Error("failed to get duel by id", zap.Error(err))
		return nil, err
	}

	return duel, nil
}
