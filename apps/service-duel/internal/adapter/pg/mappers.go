package pg

import (
	"time"

	"github.com/ccoveille/go-safecast"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg/sqlc"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
)

func mapStatus(status sqlc.DuelStatus) duelDomain.Status {
	switch status {
	case sqlc.DuelStatusWaitingForOpponent:
		return duelDomain.StatusWaitingForOpponent
	case sqlc.DuelStatusInProgress:
		return duelDomain.StatusInProgress
	case sqlc.DuelStatusCompleted:
		return duelDomain.StatusCompleted
	case sqlc.DuelStatusCancelled:
		return duelDomain.StatusCancelled
	default:
		return duelDomain.StatusWaitingForOpponent
	}
}

func mapDuel(duelRow *sqlc.Duel) (*duelDomain.Duel, error) {
	// 1. Преобразуем базовые поля из SQL-модели
	duelID, err := duelDomain.NewID(duelRow.ID.String())
	if err != nil {
		return nil, err
	}
	var winnerID *duelDomain.TelegramUserID
	if duelRow.WinnerTelegramUserID.Valid {
		id, idErr := duelDomain.NewTelegramUserID(duelRow.WinnerTelegramUserID.Int64)
		if idErr != nil {
			return nil, idErr
		}
		winnerID = &id
	}
	var nextRollDeadline *time.Time
	if duelRow.NextRollDeadline.Valid {
		nextRollDeadline = &duelRow.NextRollDeadline.Time
	}
	var completedAt *time.Time
	if duelRow.CompletedAt.Valid {
		completedAt = &duelRow.CompletedAt.Time
	}

	maxPlayersUint32, err := safecast.ToUint32(duelRow.MaxPlayers)
	if err != nil {
		return nil, err
	}
	maxGiftsUint32, err := safecast.ToUint32(duelRow.MaxGifts)
	if err != nil {
		return nil, err
	}

	maxPlayers, err := duelDomain.NewMaxPlayers(maxPlayersUint32)
	if err != nil {
		return nil, err
	}
	maxGifts, err := duelDomain.NewMaxGifts(maxGiftsUint32)
	if err != nil {
		return nil, err
	}

	params, err := duelDomain.NewParamsBuilder().
		WithIsPrivate(duelRow.IsPrivate).
		WithMaxPlayers(maxPlayers).
		WithMaxGifts(maxGifts).
		Build()
	if err != nil {
		return nil, err
	}

	// 3. Стартуем билдeр и поочерёдно заполняем все поля
	builder := duelDomain.NewDuelBuilder().
		WithID(duelID).
		WithDisplayNumber(duelRow.DisplayNumber).
		WithParams(params).
		WithStatus(mapStatus(duelRow.Status.DuelStatus)).
		WithWinner(winnerID).
		WithNextRollDeadline(nextRollDeadline).
		WithCreatedAt(duelRow.CreatedAt.Time).
		WithUpdatedAt(duelRow.UpdatedAt.Time).
		WithCompletedAt(completedAt)

	// 4. Возвращаем готовый объект
	return builder.Build()
}
