package duel

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type Repository interface {
	WithTx(tx pgx.Tx) Repository

	CreateDuel(ctx context.Context, duel *Duel) (ID, error)
	CreateStake(ctx context.Context, duelID ID, stake Stake) error
	CreateParticipant(ctx context.Context, duelID ID, participant Participant) error
	CreateRound(ctx context.Context, duelID ID, round Round) error
	CreateRoll(ctx context.Context, duelID ID, roundNumber int32, roll Roll) error

	UpdateDuelStatus(
		ctx context.Context,
		duelID ID,
		status Status,
		winnerID *TelegramUserID,
		completedAt *time.Time,
	) error
	UpdateNextRollDeadline(ctx context.Context, duelID ID, nextRollDeadline time.Time) error

	GetDuelByID(ctx context.Context, id ID) (*Duel, error)
	GetDuelList(ctx context.Context, pageRequest *shared.PageRequest) ([]*Duel, int64, error)
	FindDuelByGiftID(ctx context.Context, giftID string) (ID, error)
}
