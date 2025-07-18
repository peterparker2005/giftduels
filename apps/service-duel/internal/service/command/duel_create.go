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

type DuelCreateCommand struct {
	log               *logger.Logger
	publisher         message.Publisher
	repo              dueldomain.Repository
	txManager         pg.TxManager
	giftPrivateClient giftv1.GiftPrivateServiceClient
}

func NewDuelCreateCommand(
	repo dueldomain.Repository,
	clients *clients.Clients,
	txManager pg.TxManager,
	log *logger.Logger,
	publisher message.Publisher,
) *DuelCreateCommand {
	return &DuelCreateCommand{
		repo:              repo,
		txManager:         txManager,
		log:               log,
		giftPrivateClient: clients.Gift.Private,
		publisher:         publisher,
	}
}

type CreateDuelParams struct {
	Params dueldomain.Params
	Stakes []dueldomain.Stake
}

func (c *DuelCreateCommand) Execute(
	ctx context.Context,
	telegramUserID int64,
	params CreateDuelParams,
) (dueldomain.ID, error) {
	// 1. Начинаем транзакцию
	tx, err := c.txManager.BeginTx(ctx)
	if err != nil {
		return "", err
	}

	// 2. Подготовка defer‑компенсатора
	var (
		stakedGiftIDs []string
		duelID        dueldomain.ID
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

	// 3. Создаем доменный объект
	repo := c.repo.WithTx(tx)
	duel := dueldomain.NewDuel(params.Params)
	creator, err := dueldomain.NewParticipantBuilder().
		WithTelegramUserID(dueldomain.TelegramUserID(telegramUserID)).
		WithPhoto("").
		AsCreator().
		Build()
	if err != nil {
		return "", err
	}

	// 4. Join
	if execErr = duel.AddParticipant(creator); execErr != nil {
		return "", execErr
	}

	// 5. Резервируем ставки (StakeGift + PlaceStake)
	stakedGiftIDs, execErr = c.reserveStakes(ctx, duel, telegramUserID, params.Stakes)
	if execErr != nil {
		return "", execErr
	}

	// 6. Сохраняем все в БД (Duel, Participant, Stakes)
	duelID, execErr = c.saveDuel(ctx, repo, duel, creator, params.Stakes)
	if execErr != nil {
		return "", execErr
	}

	// 7. Коммит транзакции
	if execErr = tx.Commit(ctx); execErr != nil {
		return "", execErr
	}

	// 8. Публикация события о создании дуэли (ошибки логируем, но не откатываем)
	if pubErr := c.publishDuelCreated(duel); pubErr != nil {
		c.log.Error("failed to publish duel created", zap.Error(pubErr))
	}

	return duelID, nil
}

// reserveStakes создает все StakeGift и PlaceStake, возвращает список
// уже застейканных giftID или первую ошибку.
func (c *DuelCreateCommand) reserveStakes(
	ctx context.Context,
	duel *dueldomain.Duel,
	telegramUserID int64,
	stakes []dueldomain.Stake,
) ([]string, error) {
	var staked []string

	for i, stake := range stakes {
		c.log.Info("staking gift for duel",
			zap.String("giftID", stake.Gift.ID),
			zap.Int64("telegramUserID", telegramUserID),
		)

		resp, err := c.giftPrivateClient.StakeGift(ctx, &giftv1.StakeGiftRequest{
			GiftId:         &sharedv1.GiftId{Value: stake.Gift.ID},
			TelegramUserId: &sharedv1.TelegramUserId{Value: telegramUserID},
			GameMetadata: &giftv1.StakeGiftRequest_Duel{
				Duel: &giftv1.StakeGiftRequest_DuelMetadata{
					DuelId: &sharedv1.DuelId{Value: duel.ID.String()},
				},
			},
		})
		if err != nil {
			return staked, err
		}
		staked = append(staked, stake.Gift.ID)

		amount, err := tonamount.NewTonAmountFromString(resp.GetGift().GetPrice().GetValue())
		if err != nil {
			return staked, err
		}
		stakes[i].Gift.Price = amount
		stakes[i].TelegramUserID = dueldomain.TelegramUserID(telegramUserID)

		if err = duel.PlaceStake(stakes[i]); err != nil {
			return staked, err
		}
	}

	return staked, nil
}

// saveDuel сохраняет Duel, участника и Stakes, возвращает duelID или ошибку.
func (c *DuelCreateCommand) saveDuel(
	ctx context.Context,
	repo dueldomain.Repository,
	duel *dueldomain.Duel,
	creator dueldomain.Participant,
	stakes []dueldomain.Stake,
) (dueldomain.ID, error) {
	duelID, err := repo.CreateDuel(ctx, dueldomain.CreateDuelParams{
		DuelID:          duel.ID,
		TotalStakeValue: duel.TotalStakeValue(),
		Params:          duel.Params,
	})
	if err != nil {
		return "", err
	}

	if err = repo.CreateParticipant(ctx, duelID, creator); err != nil {
		return "", err
	}

	for _, stake := range stakes {
		if err = repo.CreateStake(ctx, duelID, stake); err != nil {
			return "", err
		}
	}

	return duelID, nil
}

func (c *DuelCreateCommand) returnStakedGifts(giftIDs []string) error {
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

func (c *DuelCreateCommand) publishDuelCreated(duel *dueldomain.Duel) error {
	event, err := proto.MapDuelCreatedEvent(duel)
	if err != nil {
		return err
	}
	data, err := googleproto.Marshal(event)
	if err != nil {
		return err
	}
	msg := message.NewMessage(duel.ID.String(), data)
	return c.publisher.Publish(duelevents.TopicDuelCreated.String(), msg)
}
