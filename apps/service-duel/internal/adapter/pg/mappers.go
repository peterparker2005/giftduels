package pg

import (
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

func mapDuel(duel *sqlc.Duel) (*duelDomain.Duel, error) {
	winnerID, err := duelDomain.NewTelegramUserID(duel.WinnerTelegramUserID.Int64)
	if err != nil {
		return nil, err
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

	return &duelDomain.Duel{
		ID:            duelID,
		DisplayNumber: duel.DisplayNumber,
		Params: duelDomain.DuelParams{
			IsPrivate:  duel.IsPrivate,
			MaxPlayers: maxPlayers,
			MaxGifts:   maxGifts,
		},
		WinnerID:         &winnerID,
		NextRollDeadline: &duel.NextRollDeadline.Time,
		TotalStakeValue:  float64(duel.TotalStakeValue.Exp),
		Status:           mapStatus(duel.Status.DuelStatus),
		CreatedAt:        duel.CreatedAt.Time,
		UpdatedAt:        duel.UpdatedAt.Time,
		CompletedAt:      &duel.CompletedAt.Time,
	}, nil
}
