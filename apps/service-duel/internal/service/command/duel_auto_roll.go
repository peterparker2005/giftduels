package command

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/proto"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	duelevents "github.com/peterparker2005/giftduels/packages/events/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
	"go.uber.org/zap"
	googleproto "google.golang.org/protobuf/proto"
)

type DuelAutoRollCommand struct {
	log                   *logger.Logger
	repo                  dueldomain.Repository
	txManager             pg.TxManager
	publisher             message.Publisher
	telegramPrivateClient telegrambotv1.TelegramBotPrivateServiceClient
}

func NewDuelAutoRollCommand(
	repo dueldomain.Repository,
	clients *clients.Clients,
	txManager pg.TxManager,
	publisher message.Publisher,
	log *logger.Logger,
) *DuelAutoRollCommand {
	return &DuelAutoRollCommand{
		repo:                  repo,
		txManager:             txManager,
		publisher:             publisher,
		log:                   log,
		telegramPrivateClient: clients.TelegramBot.Private,
	}
}

func (c *DuelAutoRollCommand) Execute(ctx context.Context, duelID dueldomain.ID) error {
	tx, err := c.txManager.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
			if err != nil {
				c.log.Error("failed to rollback tx", zap.Error(err))
			}
		}
	}()

	repo := c.repo.WithTx(tx)
	d, err := repo.GetDuelByID(ctx, duelID)
	if err != nil {
		return err
	}

	// Пока не определится единственный победитель — повторяем:
	for {
		// 1) берём текущий раунд
		round, roundErr := d.CurrentRound()
		if roundErr != nil {
			c.log.Error("failed to get current round", zap.Error(roundErr))
			// раунда может ещё не быть, но должен быть после Start() при полном составе
			return err
		}

		// 2) автокидок всех, кто ещё не кинул
		for _, pl := range round.Participants {
			if !round.HasRolled(pl) {
				val, rollErr := c.rollDice(ctx, &duelID, pl.Int64())
				if rollErr != nil {
					c.log.Error("failed to roll dice", zap.Error(rollErr))
					return rollErr
				}
				roll, rollErr := dueldomain.NewRollBuilder().
					WithTelegramUserID(pl).
					WithDiceValue(val).
					WithRolledAt(time.Now()).
					WithIsAutoRolled(true).
					Build()
				if rollErr != nil {
					c.log.Error("failed to build roll", zap.Error(rollErr))
					return rollErr
				}
				if err = d.AddRollToCurrentRound(roll); err != nil {
					c.log.Error("failed to add roll to current round", zap.Error(err))
					return err
				}
				if rollErr = repo.CreateRoll(ctx, duelID, round.RoundNumber, roll); rollErr != nil {
					c.log.Error("failed to create roll", zap.Error(rollErr))
					return rollErr
				}
			}
		}

		// 3) оцениваем раунд
		winners, finished := d.EvaluateCurrentRound()
		if !finished {
			// (маловероятно, после автокидка finished всегда true)
			next := time.Now().Add(d.TimeoutForRound())
			d.NextRollDeadline = &next
			if err = repo.UpdateNextRollDeadline(ctx, duelID, next); err != nil {
				c.log.Error("failed to update next roll deadline", zap.Error(err))
				return err
			}
			return tx.Commit(ctx)
		}

		switch len(winners) {
		case 0:
			// невозможно, но перестрахуемся
			return tx.Commit(ctx)
		case 1:
			// нашли одного — завершаем дуэль
			if err = d.Complete(winners[0]); err != nil {
				c.log.Error("failed to complete duel", zap.Error(err))
				return err
			}
			if err = repo.UpdateDuelStatus(ctx, duelID, d.Status, d.WinnerID, d.CompletedAt); err != nil {
				c.log.Error("failed to update duel status", zap.Error(err))
				return err
			}
			if err = c.sendDuelCompletedMessage(d); err != nil {
				c.log.Error("failed to send duel completed message", zap.Error(err))
				return err
			}
			return tx.Commit(ctx)
		default:
			if err = c.startNewRound(ctx, tx, d, winners); err != nil {
				return err
			}
		}
	}
}

func (c *DuelAutoRollCommand) sendDuelCompletedMessage(duel *dueldomain.Duel) error {
	msgID := uuid.New().String()
	event, err := proto.MapDuelCompletedEvent(duel)
	if err != nil {
		return err
	}
	protoBytes, err := googleproto.Marshal(event)
	if err != nil {
		return err
	}
	msg := message.NewMessage(msgID, protoBytes)

	return c.publisher.Publish(duelevents.TopicDuelCompleted.String(), msg)
}

func (c *DuelAutoRollCommand) rollDice(
	ctx context.Context,
	duelID *dueldomain.ID,
	telegramUserID int64,
) (int32, error) {
	resp, err := c.telegramPrivateClient.RollDice(ctx, &telegrambotv1.RollDiceRequest{
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

func (c *DuelAutoRollCommand) startNewRound(
	ctx context.Context,
	tx pgx.Tx,
	d *dueldomain.Duel,
	participants []dueldomain.TelegramUserID, // новые участники раунда
) error {
	repo := c.repo.WithTx(tx)
	rObj, err := dueldomain.NewRoundBuilder().
		WithRoundNumber(int32(len(d.Rounds) + 1)).
		WithParticipants(participants).
		Build()
	if err != nil {
		return err
	}
	// 1) домен
	if err = d.StartRound(participants); err != nil {
		c.log.Error("failed to start new round", zap.Error(err))
		return err
	}
	// 2) база
	if err = repo.CreateRound(ctx, d.ID, rObj); err != nil {
		c.log.Error("failed to create new round", zap.Error(err))
		return err
	}
	return nil
}
