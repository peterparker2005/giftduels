package pg

import (
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

// GiftToDomain converts sqlc.Gift to domain.Gift.
func GiftToDomain(dbGift sqlc.Gift) (*gift.Gift, error) {
	var tonAmount *tonamount.TonAmount
	if dbGift.Price.Valid {
		price, err := fromPgNumeric(dbGift.Price)
		if err != nil {
			return nil, err
		}
		tonAmount, err = tonamount.NewTonAmountFromString(price)
		if err != nil {
			return nil, err
		}
	}

	return &gift.Gift{
		ID:               pgUUIDToString(dbGift.ID),
		OwnerTelegramID:  dbGift.OwnerTelegramID,
		Status:           giftStatusToDomain(dbGift.Status),
		Price:            tonAmount,
		WithdrawnAt:      pgTimestampToTime(dbGift.WithdrawnAt),
		Title:            dbGift.Title,
		Slug:             dbGift.Slug,
		CollectibleID:    dbGift.CollectibleID,
		UpgradeMessageID: dbGift.UpgradeMessageID,
		TelegramGiftID:   dbGift.TelegramGiftID,
		CreatedAt:        pgTimestampToTimeRequired(dbGift.CreatedAt),
		UpdatedAt:        pgTimestampToTimeRequired(dbGift.UpdatedAt),
		Collection:       gift.Collection{ID: dbGift.CollectionID},
		Model:            gift.Model{ID: dbGift.ModelID},
		Backdrop:         gift.Backdrop{ID: dbGift.BackdropID},
		Symbol:           gift.Symbol{ID: dbGift.SymbolID},
		// Note: Collection, Model, Backdrop, Symbol will be populated separately.
		// since the basic GetGiftByID query doesn't include JOINs.
	}, nil
}

// GiftToDomainFromUserGiftsRow converts sqlc.Gift to domain.Gift (same as GiftToDomain).
func GiftToDomainFromUserGiftsRow(dbGift sqlc.Gift) (*gift.Gift, error) {
	return GiftToDomain(dbGift)
}

// GiftToDomainFromUserActiveGiftsRow converts sqlc.Gift to domain.Gift (same as GiftToDomain).
func GiftToDomainFromUserActiveGiftsRow(dbGift sqlc.Gift) (*gift.Gift, error) {
	return GiftToDomain(dbGift)
}

// GiftToDomainFromGiftsByIDsRow converts sqlc.Gift to domain.Gift (same as GiftToDomain).
func GiftToDomainFromGiftsByIDsRow(dbGift sqlc.Gift) (*gift.Gift, error) {
	return GiftToDomain(dbGift)
}

// GiftEventToDomain converts sqlc.GiftEvent to domain.Event.
func GiftEventToDomain(dbEvent sqlc.GiftEvent) (*gift.Event, error) {
	var relatedGameID *string
	if dbEvent.RelatedGameID.Valid {
		gameID := pgUUIDToString(dbEvent.RelatedGameID)
		relatedGameID = &gameID
	}

	telegramUserID, err := pgInt8ToInt64(dbEvent.TelegramUserID)
	if err != nil {
		return nil, err
	}

	return &gift.Event{
		ID:             pgUUIDToString(dbEvent.ID),
		GiftID:         pgUUIDToString(dbEvent.GiftID),
		TelegramUserID: telegramUserID,
		EventType:      gift.EventType(dbEvent.EventType),
		RelatedGameID:  relatedGameID,
		OccurredAt:     pgTimestampToTimeRequired(dbEvent.OccurredAt),
	}, nil
}

// CollectionToDomain converts sqlc.GiftCollection to domain.Collection.
func CollectionToDomain(dbCollection sqlc.GiftCollection) *gift.Collection {
	return &gift.Collection{
		ID:        dbCollection.ID,
		Name:      dbCollection.Name,
		ShortName: dbCollection.ShortName,
	}
}

// ModelToDomain converts sqlc.GiftModel to domain.Model.
func ModelToDomain(dbModel sqlc.GiftModel) *gift.Model {
	return &gift.Model{
		ID:             dbModel.ID,
		CollectionID:   dbModel.CollectionID,
		Name:           dbModel.Name,
		ShortName:      dbModel.ShortName,
		RarityPerMille: dbModel.RarityPerMille,
	}
}

// BackdropToDomain converts sqlc.GiftBackdrop to domain.Backdrop.
func BackdropToDomain(dbBackdrop sqlc.GiftBackdrop) *gift.Backdrop {
	return &gift.Backdrop{
		ID:             dbBackdrop.ID,
		Name:           dbBackdrop.Name,
		ShortName:      dbBackdrop.ShortName,
		RarityPerMille: dbBackdrop.RarityPerMille,
		CenterColor:    pgTextToString(dbBackdrop.CenterColor),
		EdgeColor:      pgTextToString(dbBackdrop.EdgeColor),
		PatternColor:   pgTextToString(dbBackdrop.PatternColor),
		TextColor:      pgTextToString(dbBackdrop.TextColor),
	}
}

// SymbolToDomain converts sqlc.GiftSymbol to domain.Symbol.
func SymbolToDomain(dbSymbol sqlc.GiftSymbol) *gift.Symbol {
	return &gift.Symbol{
		ID:             dbSymbol.ID,
		Name:           dbSymbol.Name,
		ShortName:      dbSymbol.ShortName,
		RarityPerMille: dbSymbol.RarityPerMille,
	}
}

// Domain to SQLC conversion functions.

// CreateGiftParamsToDB converts domain.CreateGiftParams to sqlc.CreateGiftParams.
func CreateGiftParamsToDB(params *gift.CreateGiftParams) (sqlc.CreateGiftParams, error) {
	if params == nil {
		return sqlc.CreateGiftParams{}, errors.New("params cannot be nil")
	}

	var pgPrice pgtype.Numeric
	if params.Price != nil {
		var err error
		pgPrice, err = pgNumeric(params.Price.String())
		if err != nil {
			return sqlc.CreateGiftParams{}, err
		}
	}

	return sqlc.CreateGiftParams{
		ID:               mustPgUUID(params.GiftID),
		TelegramGiftID:   params.TelegramGiftID,
		Title:            params.Title,
		Slug:             params.Slug,
		OwnerTelegramID:  params.OwnerTelegramID,
		UpgradeMessageID: params.UpgradeMessageID,
		Price:            pgPrice,
		CollectibleID:    params.CollectibleID,
		Status:           giftStatusToSQLC(params.Status),
		CreatedAt:        timeToPgTimestamp(params.CreatedAt),
		UpdatedAt:        timeToPgTimestamp(params.UpdatedAt),
	}, nil
}

// CreateGiftEventParamsToDB converts domain.CreateGiftEventParams to sqlc.CreateGiftEventParams.
func CreateGiftEventParamsToDB(
	params gift.CreateGiftEventParams,
) (sqlc.CreateGiftEventParams, error) {
	telegramUserID := int64PtrToPgInt8(&params.TelegramUserID)
	var relatedGameID pgtype.UUID
	if params.RelatedGameID != nil {
		var err error
		relatedGameID, err = pgUUID(*params.RelatedGameID)
		if err != nil {
			return sqlc.CreateGiftEventParams{}, err
		}
	}
	return sqlc.CreateGiftEventParams{
		GiftID:         mustPgUUID(params.GiftID),
		TelegramUserID: telegramUserID,
		EventType:      sqlc.GiftEventType(params.EventType),
		RelatedGameID:  relatedGameID,
	}, nil
}

// CreateCollectionParamsToDB converts domain.CreateCollectionParams to sqlc.CreateCollectionParams.
func CreateCollectionParamsToDB(params *gift.CreateCollectionParams) sqlc.CreateCollectionParams {
	if params == nil {
		return sqlc.CreateCollectionParams{}
	}
	return sqlc.CreateCollectionParams{
		Name:      params.Name,
		ShortName: params.ShortName,
	}
}

// CreateModelParamsToDB converts domain.CreateModelParams to sqlc.CreateModelParams.
func CreateModelParamsToDB(params *gift.CreateModelParams) sqlc.CreateModelParams {
	if params == nil {
		return sqlc.CreateModelParams{}
	}
	return sqlc.CreateModelParams{
		CollectionID:   params.CollectionID,
		Name:           params.Name,
		ShortName:      params.ShortName,
		RarityPerMille: params.RarityPerMille,
	}
}

// CreateBackdropParamsToDB converts domain.CreateBackdropParams to sqlc.CreateBackdropParams.
func CreateBackdropParamsToDB(params *gift.CreateBackdropParams) sqlc.CreateBackdropParams {
	if params == nil {
		return sqlc.CreateBackdropParams{}
	}
	return sqlc.CreateBackdropParams{
		Name:           params.Name,
		ShortName:      params.ShortName,
		RarityPerMille: params.RarityPerMille,
		CenterColor:    stringPtrToPgText(params.CenterColor),
		EdgeColor:      stringPtrToPgText(params.EdgeColor),
		PatternColor:   stringPtrToPgText(params.PatternColor),
		TextColor:      stringPtrToPgText(params.TextColor),
	}
}

// CreateSymbolParamsToDB converts domain.CreateSymbolParams to sqlc.CreateSymbolParams.
func CreateSymbolParamsToDB(params *gift.CreateSymbolParams) sqlc.CreateSymbolParams {
	if params == nil {
		return sqlc.CreateSymbolParams{}
	}
	return sqlc.CreateSymbolParams{
		Name:           params.Name,
		ShortName:      params.ShortName,
		RarityPerMille: params.RarityPerMille,
	}
}

// Status conversion functions.

// giftStatusToDomain converts sqlc.GiftStatus to domain.Status.
func giftStatusToDomain(status sqlc.GiftStatus) gift.Status {
	switch status {
	case sqlc.GiftStatusOwned:
		return gift.StatusOwned
	case sqlc.GiftStatusInGame:
		return gift.StatusInGame
	case sqlc.GiftStatusWithdrawPending:
		return gift.StatusWithdrawPending
	case sqlc.GiftStatusWithdrawn:
		return gift.StatusWithdrawn
	default:
		// Log unknown status and return default
		return gift.StatusOwned // fallback to safe default
	}
}

// giftStatusToSQLC converts domain.Status to sqlc.GiftStatus.
func giftStatusToSQLC(status gift.Status) sqlc.GiftStatus {
	switch status {
	case gift.StatusOwned:
		return sqlc.GiftStatusOwned
	case gift.StatusInGame:
		return sqlc.GiftStatusInGame
	case gift.StatusWithdrawPending:
		return sqlc.GiftStatusWithdrawPending
	case gift.StatusWithdrawn:
		return sqlc.GiftStatusWithdrawn
	default:
		// Log unknown status and return default
		return sqlc.GiftStatusOwned // fallback to safe default
	}
}
