package pg

import (
	"time"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg/sqlc"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
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

func mapDuel(duel *sqlc.Duel) (*duelDomain.Duel, error) {
	var winnerID *duelDomain.TelegramUserID
	if duel.WinnerTelegramUserID.Valid {
		id, err := duelDomain.NewTelegramUserID(duel.WinnerTelegramUserID.Int64)
		if err != nil {
			return nil, err
		}
		winnerID = &id
	}

	duelID, err := duelDomain.NewID(duel.ID.String())
	if err != nil {
		return nil, err
	}

	maxPlayers, err := duelDomain.NewMaxPlayers(duel.MaxPlayers)
	if err != nil {
		return nil, err
	}

	maxGifts, err := duelDomain.NewMaxGifts(duel.MaxGifts)
	if err != nil {
		return nil, err
	}

	totalStakeValueStr, err := fromPgNumeric(duel.TotalStakeValue)
	if err != nil {
		return nil, err
	}

	totalStakeValue, err := tonamount.NewTonAmountFromString(totalStakeValueStr)
	if err != nil {
		return nil, err
	}

	var nextRollDeadline *time.Time
	if duel.NextRollDeadline.Valid {
		nextRollDeadline = &duel.NextRollDeadline.Time
	}

	var completedAt *time.Time
	if duel.CompletedAt.Valid {
		completedAt = &duel.CompletedAt.Time
	}

	return &duelDomain.Duel{
		ID:            duelID,
		DisplayNumber: duel.DisplayNumber,
		Params: duelDomain.Params{
			IsPrivate:  duel.IsPrivate,
			MaxPlayers: maxPlayers,
			MaxGifts:   maxGifts,
		},
		WinnerID:         winnerID,
		NextRollDeadline: nextRollDeadline,
		TotalStakeValue:  totalStakeValue,
		Status:           mapStatus(duel.Status.DuelStatus),
		CreatedAt:        duel.CreatedAt.Time,
		UpdatedAt:        duel.UpdatedAt.Time,
		CompletedAt:      completedAt,
	}, nil
}
