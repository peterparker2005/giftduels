package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/adapter/pg/sqlc"
	duelDomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/shared"
)

type duelRepository struct {
	q      *sqlc.Queries
	logger *logger.Logger
}

func NewDuelRepository(pool *pgxpool.Pool, logger *logger.Logger) duelDomain.Repository {
	return &duelRepository{q: sqlc.New(pool), logger: logger}
}

func (r *duelRepository) WithTx(tx pgx.Tx) duelDomain.Repository {
	return &duelRepository{q: r.q.WithTx(tx), logger: r.logger}
}

func (r *duelRepository) CreateDuel(ctx context.Context, params duelDomain.CreateDuelParams) (duelDomain.ID, error) {
	duel, err := r.q.CreateDuel(ctx, sqlc.CreateDuelParams{
		IsPrivate:  params.Params.IsPrivate,
		MaxPlayers: int32(params.Params.MaxPlayers),
		MaxGifts:   int32(params.Params.MaxGifts),
	})
	if err != nil {
		return "", err
	}

	return duelDomain.NewID(duel.ID.String())
}

func (r *duelRepository) GetDuelByID(ctx context.Context, id duelDomain.ID) (*duelDomain.Duel, error) {
	sqlcDuel, err := r.q.GetDuelByID(ctx, mustPgUUID(id.String()))
	if err != nil {
		return nil, err
	}

	duel, err := mapDuel(&sqlcDuel)
	if err != nil {
		return nil, err
	}

	return duel, nil
}

func (r *duelRepository) GetDuelList(ctx context.Context, pageRequest *shared.PageRequest) ([]*duelDomain.Duel, int64, error) {
	total, err := r.q.GetVisibleDuelsCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	sqlcDuels, err := r.q.GetDuels(ctx, sqlc.GetDuelsParams{
		Limit:  pageRequest.PageSize(),
		Offset: pageRequest.Offset(),
	})
	if err != nil {
		return nil, 0, err
	}

	duels := make([]*duelDomain.Duel, len(sqlcDuels))
	for i, sqlcDuel := range sqlcDuels {
		duel, err := mapDuel(&sqlcDuel)
		if err != nil {
			return nil, 0, err
		}
		duels[i] = duel
	}

	return duels, total, nil
}
