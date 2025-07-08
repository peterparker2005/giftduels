package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/zap"
)

type GiftRepository struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	logger *logger.Logger
}

func NewGiftRepo(pool *pgxpool.Pool, logger *logger.Logger) gift.GiftRepository {
	return &GiftRepository{pool: pool, q: sqlc.New(pool), logger: logger}
}

func (r *GiftRepository) GetGiftByID(ctx context.Context, id string) (*gift.Gift, error) {
	dbGift, err := r.q.GetGiftByID(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}
	return GiftToDomain(dbGift), nil
}

func (r *GiftRepository) GetUserGifts(ctx context.Context, limit, offset int32, ownerTelegramID int64) (*gift.GetUserGiftsResult, error) {
	total, err := r.q.GetUserGiftsCount(ctx, ownerTelegramID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.GetUserGifts(ctx, sqlc.GetUserGiftsParams{
		Limit:           limit,
		Offset:          offset,
		OwnerTelegramID: ownerTelegramID,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*gift.Gift, len(rows))
	for i, row := range rows {
		out[i] = GiftToDomain(row)
	}
	return &gift.GetUserGiftsResult{
		Gifts: out,
		Total: total,
	}, nil
}

func (r *GiftRepository) StakeGiftForGame(ctx context.Context, id string) (*gift.Gift, error) {
	dbGift, err := r.q.StakeGiftForGame(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}
	return GiftToDomain(dbGift), nil
}

func (r *GiftRepository) UpdateGiftOwner(ctx context.Context, id string, ownerTelegramID int64) (*gift.Gift, error) {
	dbGift, err := r.q.UpdateGiftOwner(ctx, sqlc.UpdateGiftOwnerParams{
		ID:              mustPgUUID(id),
		OwnerTelegramID: ownerTelegramID,
	})
	if err != nil {
		return nil, err
	}
	return GiftToDomain(dbGift), nil
}

func (r *GiftRepository) MarkGiftForWithdrawal(ctx context.Context, id string) (*gift.Gift, error) {
	dbGift, err := r.q.MarkGiftForWithdrawal(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}
	return GiftToDomain(dbGift), nil
}

func (r *GiftRepository) CompleteGiftWithdrawal(ctx context.Context, id string) (*gift.Gift, error) {
	dbGift, err := r.q.CompleteGiftWithdrawal(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}
	return GiftToDomain(dbGift), nil
}

func (r *GiftRepository) CreateGiftEvent(ctx context.Context, giftID string, fromUserID, toUserID int64) (*gift.GiftEvent, error) {
	return nil, nil
}

func (r *GiftRepository) GetGiftEvents(ctx context.Context, giftID string, limit int32, offset int32) ([]*gift.GiftEvent, error) {
	return nil, nil
}

func (r *GiftRepository) CreateGiftWithDetails(
	ctx context.Context,
	in *gift.CreateGiftParams,
	attrs []gift.CreateGiftAttributeParams,
) (*gift.Gift, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				r.logger.Error("rollback", zap.Error(err))
			}
		}
	}()

	qtx := r.q.WithTx(tx)

	dbGift, err := qtx.CreateGift(ctx, CreateGiftParamsToDB(in))
	if err != nil {
		return nil, err
	}

	for _, a := range attrs {
		if _, err = qtx.CreateGiftAttribute(ctx, CreateGiftAttributeParamsToDB(&a)); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return GiftToDomain(dbGift), nil
}
