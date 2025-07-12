package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
)

type GiftRepository struct {
	pool   *pgxpool.Pool
	q      *sqlc.Queries
	logger *logger.Logger
}

func NewGiftRepo(pool *pgxpool.Pool, logger *logger.Logger) gift.GiftRepository {
	return &GiftRepository{pool: pool, q: sqlc.New(pool), logger: logger}
}

func (r *GiftRepository) WithTx(tx pgx.Tx) gift.GiftRepository {
	return &GiftRepository{pool: r.pool, q: r.q.WithTx(tx), logger: r.logger}
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
		out[i] = GiftToDomainFromUserGiftsRow(row)
	}
	return &gift.GetUserGiftsResult{
		Gifts: out,
		Total: total,
	}, nil
}

func (r *GiftRepository) GetUserActiveGifts(ctx context.Context, limit, offset int32, ownerTelegramID int64) (*gift.GetUserGiftsResult, error) {
	total, err := r.q.GetUserActiveGiftsCount(ctx, ownerTelegramID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.GetUserActiveGifts(ctx, sqlc.GetUserActiveGiftsParams{
		Limit:           limit,
		Offset:          offset,
		OwnerTelegramID: ownerTelegramID,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*gift.Gift, len(rows))
	for i, row := range rows {
		out[i] = GiftToDomainFromUserActiveGiftsRow(row)
	}
	return &gift.GetUserGiftsResult{
		Gifts: out,
		Total: total,
	}, nil
}

func (r *GiftRepository) StakeGiftForGame(ctx context.Context, id string) (*gift.Gift, error) {
	_, err := r.q.StakeGiftForGame(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}

	// Get full gift details with joins
	return r.GetGiftByID(ctx, id)
}

func (r *GiftRepository) UpdateGiftOwner(ctx context.Context, id string, ownerTelegramID int64) (*gift.Gift, error) {
	_, err := r.q.UpdateGiftOwner(ctx, sqlc.UpdateGiftOwnerParams{
		ID:              mustPgUUID(id),
		OwnerTelegramID: ownerTelegramID,
	})
	if err != nil {
		return nil, err
	}

	// Get full gift details with joins
	return r.GetGiftByID(ctx, id)
}

func (r *GiftRepository) MarkGiftForWithdrawal(ctx context.Context, id string) (*gift.Gift, error) {
	_, err := r.q.MarkGiftForWithdrawal(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}

	// Get full gift details with joins
	return r.GetGiftByID(ctx, id)
}

func (r *GiftRepository) CancelGiftWithdrawal(ctx context.Context, id string) (*gift.Gift, error) {
	_, err := r.q.CancelGiftWithdrawal(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}

	// Get full gift details with joins
	return r.GetGiftByID(ctx, id)
}

func (r *GiftRepository) CompleteGiftWithdrawal(ctx context.Context, id string) (*gift.Gift, error) {
	_, err := r.q.CompleteGiftWithdrawal(ctx, mustPgUUID(id))
	if err != nil {
		return nil, err
	}

	// Get full gift details with joins
	return r.GetGiftByID(ctx, id)
}

func (r *GiftRepository) CreateGiftEvent(ctx context.Context, params gift.CreateGiftEventParams) (*gift.GiftEvent, error) {
	dbGiftEvent, err := r.q.CreateGiftEvent(ctx, CreateGiftEventParamsToDB(params))
	if err != nil {
		return nil, err
	}
	return GiftEventToDomain(dbGiftEvent), nil
}

func (r *GiftRepository) GetGiftEvents(ctx context.Context, giftID string, limit int32, offset int32) ([]*gift.GiftEvent, error) {
	dbEvents, err := r.q.GetGiftEvents(ctx, sqlc.GetGiftEventsParams{
		GiftID: mustPgUUID(giftID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	events := make([]*gift.GiftEvent, len(dbEvents))
	for i, dbEvent := range dbEvents {
		events[i] = GiftEventToDomain(dbEvent)
	}
	return events, nil
}

func (r *GiftRepository) CreateGift(ctx context.Context, params *gift.CreateGiftParams, collectionID, modelID, backdropID, symbolID int32) (*gift.Gift, error) {
	createParams := CreateGiftParamsToDB(params)
	createParams.CollectionID = collectionID
	createParams.ModelID = modelID
	createParams.BackdropID = backdropID
	createParams.SymbolID = symbolID

	_, err := r.q.CreateGift(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("create gift: %w", err)
	}

	// Get the created gift with full details
	return r.GetGiftByID(ctx, params.GiftID)
}

func (r *GiftRepository) GetGiftsByIDs(ctx context.Context, ids []string) ([]*gift.Gift, error) {
	pgUUIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgUUIDs[i] = mustPgUUID(id)
	}

	rows, err := r.q.GetGiftsByIDs(ctx, pgUUIDs)
	if err != nil {
		return nil, err
	}

	gifts := make([]*gift.Gift, len(rows))
	for i, row := range rows {
		gifts[i] = GiftToDomainFromGiftsByIDsRow(row)
	}
	return gifts, nil
}

func (r *GiftRepository) SaveGiftWithPrice(ctx context.Context, id string, price float64) (*gift.Gift, error) {
	_, err := r.q.SaveGiftWithPrice(ctx, sqlc.SaveGiftWithPriceParams{
		ID:    mustPgUUID(id),
		Price: price,
	})
	if err != nil {
		return nil, err
	}

	// Get full gift details with joins
	return r.GetGiftByID(ctx, id)
}

// Lookup table methods
func (r *GiftRepository) GetGiftModel(ctx context.Context, id int32) (*gift.Model, error) {
	dbModel, err := r.q.GetGiftModel(ctx, id)
	if err != nil {
		return nil, err
	}
	return ModelToDomain(dbModel), nil
}

func (r *GiftRepository) GetGiftBackdrop(ctx context.Context, id int32) (*gift.Backdrop, error) {
	dbBackdrop, err := r.q.GetGiftBackdrop(ctx, id)
	if err != nil {
		return nil, err
	}
	return BackdropToDomain(dbBackdrop), nil
}

func (r *GiftRepository) GetGiftSymbol(ctx context.Context, id int32) (*gift.Symbol, error) {
	dbSymbol, err := r.q.GetGiftSymbol(ctx, id)
	if err != nil {
		return nil, err
	}
	return SymbolToDomain(dbSymbol), nil
}

// Create lookup table methods
func (r *GiftRepository) CreateCollection(ctx context.Context, params *gift.CreateCollectionParams) (*gift.Collection, error) {
	dbCollection, err := r.q.CreateCollection(ctx, CreateCollectionParamsToDB(params))
	if err != nil {
		return nil, err
	}
	return &gift.Collection{
		ID:        dbCollection.ID,
		Name:      dbCollection.Name,
		ShortName: dbCollection.ShortName,
	}, nil
}

func (r *GiftRepository) CreateModel(ctx context.Context, params *gift.CreateModelParams) (*gift.Model, error) {
	dbModel, err := r.q.CreateModel(ctx, CreateModelParamsToDB(params))
	if err != nil {
		return nil, err
	}
	return ModelToDomain(dbModel), nil
}

func (r *GiftRepository) CreateBackdrop(ctx context.Context, params *gift.CreateBackdropParams) (*gift.Backdrop, error) {
	dbBackdrop, err := r.q.CreateBackdrop(ctx, CreateBackdropParamsToDB(params))
	if err != nil {
		return nil, err
	}
	return BackdropToDomain(dbBackdrop), nil
}

func (r *GiftRepository) CreateSymbol(ctx context.Context, params *gift.CreateSymbolParams) (*gift.Symbol, error) {
	dbSymbol, err := r.q.CreateSymbol(ctx, CreateSymbolParamsToDB(params))
	if err != nil {
		return nil, err
	}
	return SymbolToDomain(dbSymbol), nil
}

// Find lookup table methods
func (r *GiftRepository) FindCollectionByName(ctx context.Context, name string) (*gift.Collection, error) {
	dbCollection, err := r.q.FindCollectionByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &gift.Collection{
		ID:        dbCollection.ID,
		Name:      dbCollection.Name,
		ShortName: dbCollection.ShortName,
	}, nil
}

func (r *GiftRepository) FindModelByName(ctx context.Context, name string) (*gift.Model, error) {
	dbModel, err := r.q.FindModelByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return ModelToDomain(dbModel), nil
}

func (r *GiftRepository) FindBackdropByName(ctx context.Context, name string) (*gift.Backdrop, error) {
	dbBackdrop, err := r.q.FindBackdropByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return BackdropToDomain(dbBackdrop), nil
}

func (r *GiftRepository) FindSymbolByName(ctx context.Context, name string) (*gift.Symbol, error) {
	dbSymbol, err := r.q.FindSymbolByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return SymbolToDomain(dbSymbol), nil
}
