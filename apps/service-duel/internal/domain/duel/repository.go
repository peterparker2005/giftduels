package duel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type CreateDuelParams struct {
	DuelID       ID
	Params       Params
	Participants []Participant
	Stakes       []Stake
}

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	CreateDuel(ctx context.Context, params CreateDuelParams) (ID, error)
	GetDuelByID(ctx context.Context, id ID) (*Duel, error)
	GetDuelList(ctx context.Context, pageRequest *shared.PageRequest) ([]*Duel, int64, error)
}
