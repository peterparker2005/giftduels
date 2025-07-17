package command

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	telegrambotv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/telegrambot/v1"
)

type DuelAutoRollCommand struct {
	log                   *logger.Logger
	repo                  dueldomain.Repository
	txManager             pg.TxManager
	telegramPrivateClient telegrambotv1.TelegramBotPrivateServiceClient
}

func NewDuelAutoRollCommand(
	repo dueldomain.Repository,
	clients *clients.Clients,
	txManager pg.TxManager,
	log *logger.Logger,
) *DuelAutoRollCommand {
	return &DuelAutoRollCommand{
		repo:                  repo,
		txManager:             txManager,
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
			tx.Rollback(ctx)
		}
	}()
	repo := c.repo.WithTx(tx)

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
			val, err := c.rollDice(ctx, &duelID, pl.Int64())
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
		if err := c.startNewRound(ctx, tx, d); err != nil {
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

func (c *DuelAutoRollCommand) startNewRound(ctx context.Context, tx pgx.Tx, d *dueldomain.Duel) error {
	repo := c.repo.WithTx(tx)
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
