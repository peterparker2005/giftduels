package command

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/proto"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelevents "github.com/peterparker2005/giftduels/packages/events/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
	googleproto "google.golang.org/protobuf/proto"
)

type DuelJoinCommand struct {
	log               *logger.Logger
	publisher         message.Publisher
	repo              dueldomain.Repository
	txManager         pg.TxManager
	giftPrivateClient giftv1.GiftPrivateServiceClient
	scheduler         dueldomain.Scheduler
}

func NewDuelJoinCommand(
	repo dueldomain.Repository,
	clients *clients.Clients,
	txManager pg.TxManager,
	log *logger.Logger,
	publisher message.Publisher,
	scheduler dueldomain.Scheduler,
) *DuelJoinCommand {
	return &DuelJoinCommand{
		repo:              repo,
		txManager:         txManager,
		log:               log,
		giftPrivateClient: clients.Gift.Private,
		publisher:         publisher,
		scheduler:         scheduler,
	}
}

func (c *DuelJoinCommand) Execute(
	ctx context.Context,
	duelID dueldomain.ID,
	giftIDs []string,
	telegramUserIDInt64 int64,
) error {
	// 1. Начинаем транзакцию
	tx, err := c.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}

	// 2. Подготовка defer‑компенсатора
	var (
		stakedGiftIDs []string
		execErr       error
	)
	defer func() {
		if execErr != nil {
			if len(stakedGiftIDs) > 0 {
				if compErr := c.returnStakedGifts(stakedGiftIDs); compErr != nil {
					c.log.Error("compensation failed", zap.Error(compErr))
				}
			}
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				c.log.Error("rollback failed", zap.Error(rbErr))
			}
		}
	}()

	repo := c.repo.WithTx(tx)

	// 3. Получаем дуэль из БД
	duel, err := repo.GetDuelByID(ctx, duelID)
	if err != nil {
		execErr = err
		return err
	}

	// 3.1. Загружаем цены гифтов из gift service для валидации entry price range
	if err = c.loadGiftPrices(ctx, duel); err != nil {
		execErr = err
		return err
	}

	// 4. Создаем TelegramUserID
	telegramUserID, err := dueldomain.NewTelegramUserID(telegramUserIDInt64)
	if err != nil {
		execErr = err
		return err
	}

	// 5. Создаем участника (не создатель)
	participant, err := dueldomain.NewParticipantBuilder().
		WithTelegramUserID(telegramUserID).
		Build()
	if err != nil {
		return err
	}

	// 6. Присоединяемся к дуэли
	if execErr = duel.AddParticipant(participant); execErr != nil {
		return execErr
	}

	// 7. Резервируем ставки (StakeGift + PlaceStake)
	stakes, stakedGiftIDs, execErr := c.reserveStakes(ctx, duel, telegramUserIDInt64, giftIDs)
	if execErr != nil {
		return execErr
	}

	// 8. Проверяем entry price range
	if execErr = duel.ValidateEntry(telegramUserID); execErr != nil {
		return execErr
	}

	// 9. Сохраняем участника в БД
	if execErr = repo.CreateParticipant(ctx, duelID, participant); execErr != nil {
		return execErr
	}

	// 10. Сохраняем ставки в БД
	for _, stake := range stakes {
		if execErr = repo.CreateStake(ctx, duelID, stake); execErr != nil {
			return execErr
		}
	}

	// 11. Проверяем, нужно ли запускать дуэль
	if len(duel.Participants) == int(duel.Params.MaxPlayers) {
		if execErr = duel.Start(); execErr != nil {
			return execErr
		}
		// сохраняем новый раунд
		if execErr = repo.CreateRound(ctx, duelID, duel.Rounds[len(duel.Rounds)-1]); execErr != nil {
			return execErr
		}
		// меняем статус и дедлайн в таблице duels
		if execErr = repo.UpdateDuelStatus(ctx, duelID, duel.Status, nil, nil); execErr != nil {
			return execErr
		}
		if execErr = repo.UpdateNextRollDeadline(ctx, duelID, *duel.NextRollDeadline); execErr != nil {
			return execErr
		}
	} else {
		// Если дуэль еще не заполнена, не планируем авто-бросок
		return tx.Commit(ctx)
	}

	// 11. Планируем авто-бросок
	if execErr = c.scheduler.ScheduleAutoRoll(duelID, *duel.NextRollDeadline); execErr != nil {
		return execErr
	}

	// 12. Коммитим транзакцию
	if execErr = tx.Commit(ctx); execErr != nil {
		return execErr
	}

	// 13. Публикация события о присоединении к дуэли (ошибки логируем, но не откатываем)
	if pubErr := c.publishDuelJoined(duel, telegramUserID); pubErr != nil {
		c.log.Error("failed to publish duel joined", zap.Error(pubErr))
	}

	return nil
}

// reserveStakes создает все StakeGift и PlaceStake, возвращает список
// ставок, уже застейканных giftID или первую ошибку.
func (c *DuelJoinCommand) reserveStakes(
	ctx context.Context,
	duel *dueldomain.Duel,
	telegramUserIDInt64 int64,
	giftIDs []string,
) ([]dueldomain.Stake, []string, error) {
	var stakes []dueldomain.Stake
	var staked []string

	for _, giftID := range giftIDs {
		c.log.Info("staking gift for duel join",
			zap.String("giftID", giftID),
			zap.Int64("telegramUserID", telegramUserIDInt64),
		)

		resp, err := c.giftPrivateClient.StakeGift(ctx, &giftv1.StakeGiftRequest{
			GiftId:         &sharedv1.GiftId{Value: giftID},
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserIDInt64},
			GameMetadata: &giftv1.StakeGiftRequest_Duel{
				Duel: &giftv1.StakeGiftRequest_DuelMetadata{
					DuelId: &sharedv1.DuelId{Value: duel.ID.String()},
				},
			},
		})
		if err != nil {
			return stakes, staked, err
		}
		staked = append(staked, giftID)

		amount, err := tonamount.NewTonAmountFromString(resp.GetGift().GetPrice().GetValue())
		if err != nil {
			return stakes, staked, err
		}

		gift, err := dueldomain.NewStakedGiftBuilder().
			WithID(giftID).
			WithTitle(resp.GetGift().GetTitle()).
			WithSlug(resp.GetGift().GetSlug()).
			WithPrice(amount).
			Build()
		if err != nil {
			return stakes, staked, err
		}

		stake, stakeErr := dueldomain.NewStakeBuilder(dueldomain.TelegramUserID(telegramUserIDInt64)).
			WithGift(gift).
			Build()
		if stakeErr != nil {
			return stakes, staked, stakeErr
		}

		if err = duel.PlaceStake(stake); err != nil {
			return stakes, staked, err
		}

		stakes = append(stakes, stake)
	}

	return stakes, staked, nil
}

func (c *DuelJoinCommand) returnStakedGifts(giftIDs []string) error {
	for _, giftID := range giftIDs {
		c.log.Info("returning gift from game due to error", zap.String("giftID", giftID))
		msg := message.NewMessage(uuid.New().String(), []byte(giftID))
		if err := c.publisher.Publish(duelevents.TopicDuelCreateFailed.String(), msg); err != nil {
			c.log.Error(
				"failed to return gift during cleanup",
				zap.String("giftID", giftID),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func (c *DuelJoinCommand) loadGiftPrices(ctx context.Context, duel *dueldomain.Duel) error {
	// Загружаем цены гифтов из gift service для всех ставок
	for i, stake := range duel.Stakes {
		giftResp, err := c.giftPrivateClient.PrivateGetGift(ctx, &giftv1.PrivateGetGiftRequest{
			GiftId: &sharedv1.GiftId{Value: stake.Gift.ID},
		})
		if err != nil {
			c.log.Error("failed to get gift from gift service",
				zap.String("giftID", stake.Gift.ID),
				zap.Error(err))
			return err
		}

		// Parse price from gift service response
		price, err := tonamount.NewTonAmountFromString(giftResp.GetGift().GetPrice().GetValue())
		if err != nil {
			c.log.Error("failed to parse gift price",
				zap.String("giftID", stake.Gift.ID),
				zap.String("price", giftResp.GetGift().GetPrice().GetValue()),
				zap.Error(err))
			return err
		}

		// Update the gift with price and other details
		gift, err := dueldomain.NewStakedGiftBuilder().
			WithID(stake.Gift.ID).
			WithTitle(giftResp.GetGift().GetTitle()).
			WithSlug(giftResp.GetGift().GetSlug()).
			WithPrice(price).
			Build()
		if err != nil {
			c.log.Error("failed to build staked gift", zap.Error(err))
			return err
		}

		duel.Stakes[i].Gift = gift
	}

	return nil
}

func (c *DuelJoinCommand) publishDuelJoined(
	duel *dueldomain.Duel,
	userID dueldomain.TelegramUserID,
) error {
	event, err := proto.MapDuelJoinedEvent(duel, userID)
	if err != nil {
		return err
	}
	data, err := googleproto.Marshal(event)
	if err != nil {
		return err
	}
	msg := message.NewMessage(duel.ID.String(), data)
	return c.publisher.Publish(duelevents.TopicDuelJoined.String(), msg)
}
