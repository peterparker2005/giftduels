package pg

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg/sqlc"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
)

// GiftToDomain преобразует PG модель в domain модель
func GiftToDomain(dbGift sqlc.GetGiftByIDRow) *gift.Gift {
	var withdrawnAt *time.Time
	if dbGift.WithdrawnAt.Valid {
		withdrawnAt = &dbGift.WithdrawnAt.Time
	}

	return &gift.Gift{
		ID:               dbGift.ID.String(),
		OwnerTelegramID:  dbGift.OwnerTelegramID,
		Status:           gift.Status(dbGift.Status),
		Price:            dbGift.Price,
		WithdrawnAt:      withdrawnAt,
		Title:            dbGift.Title,
		Slug:             dbGift.Slug,
		CollectibleID:    dbGift.CollectibleID,
		UpgradeMessageID: dbGift.UpgradeMessageID,
		TelegramGiftID:   dbGift.TelegramGiftID,
		CreatedAt:        dbGift.CreatedAt.Time,
		UpdatedAt:        dbGift.UpdatedAt.Time,
		Collection: gift.Collection{
			ID:        dbGift.CollectionID,
			Name:      dbGift.CollectionName,
			ShortName: dbGift.CollectionShortName,
		},
		Model: gift.Model{
			ID:             dbGift.ModelID,
			CollectionID:   dbGift.CollectionID,
			Name:           dbGift.ModelName,
			ShortName:      dbGift.ModelShortName,
			RarityPerMille: dbGift.ModelRarity,
		},
		Backdrop: gift.Backdrop{
			ID:             dbGift.BackdropID,
			Name:           dbGift.BackdropName,
			ShortName:      dbGift.BackdropShortName,
			RarityPerMille: dbGift.BackdropRarity,
			CenterColor:    stringPtrFromPgText(dbGift.CenterColor),
			EdgeColor:      stringPtrFromPgText(dbGift.EdgeColor),
			PatternColor:   stringPtrFromPgText(dbGift.PatternColor),
			TextColor:      stringPtrFromPgText(dbGift.TextColor),
		},
		Symbol: gift.Symbol{
			ID:             dbGift.SymbolID,
			Name:           dbGift.SymbolName,
			ShortName:      dbGift.SymbolShortName,
			RarityPerMille: dbGift.SymbolRarity,
		},
	}
}

// GiftToDomainFromUserGiftsRow преобразует PG модель из GetUserGiftsRow в domain модель
func GiftToDomainFromUserGiftsRow(dbGift sqlc.GetUserGiftsRow) *gift.Gift {
	var withdrawnAt *time.Time
	if dbGift.WithdrawnAt.Valid {
		withdrawnAt = &dbGift.WithdrawnAt.Time
	}

	return &gift.Gift{
		ID:               dbGift.ID.String(),
		OwnerTelegramID:  dbGift.OwnerTelegramID,
		Status:           gift.Status(dbGift.Status),
		Price:            dbGift.Price,
		WithdrawnAt:      withdrawnAt,
		Title:            dbGift.Title,
		Slug:             dbGift.Slug,
		CollectibleID:    dbGift.CollectibleID,
		UpgradeMessageID: dbGift.UpgradeMessageID,
		TelegramGiftID:   dbGift.TelegramGiftID,
		CreatedAt:        dbGift.CreatedAt.Time,
		UpdatedAt:        dbGift.UpdatedAt.Time,
		Collection: gift.Collection{
			ID:        dbGift.CollectionID,
			Name:      dbGift.CollectionName,
			ShortName: dbGift.CollectionShortName,
		},
		Model: gift.Model{
			ID:             dbGift.ModelID,
			CollectionID:   dbGift.CollectionID,
			Name:           dbGift.ModelName,
			ShortName:      dbGift.ModelShortName,
			RarityPerMille: 0, // Not included in GetUserGifts query
		},
		Backdrop: gift.Backdrop{
			ID:             dbGift.BackdropID,
			Name:           dbGift.BackdropName,
			ShortName:      dbGift.BackdropShortName,
			RarityPerMille: 0, // Not included in GetUserGifts query
		},
		Symbol: gift.Symbol{
			ID:             dbGift.SymbolID,
			Name:           dbGift.SymbolName,
			ShortName:      dbGift.SymbolShortName,
			RarityPerMille: 0, // Not included in GetUserGifts query
		},
	}
}

// GiftToDomainFromUserActiveGiftsRow преобразует PG модель из GetUserActiveGiftsRow в domain модель
func GiftToDomainFromUserActiveGiftsRow(db sqlc.GetUserActiveGiftsRow) *gift.Gift {
	var withdrawnAt *time.Time
	if db.WithdrawnAt.Valid {
		withdrawnAt = &db.WithdrawnAt.Time
	}

	return &gift.Gift{
		ID:               db.ID.String(),
		OwnerTelegramID:  db.OwnerTelegramID,
		Status:           gift.Status(db.Status),
		Price:            db.Price,
		WithdrawnAt:      withdrawnAt,
		Title:            db.Title,
		Slug:             db.Slug,
		CollectibleID:    db.CollectibleID,
		UpgradeMessageID: db.UpgradeMessageID,
		TelegramGiftID:   db.TelegramGiftID,
		CreatedAt:        db.CreatedAt.Time,
		UpdatedAt:        db.UpdatedAt.Time,

		Collection: gift.Collection{
			ID:        db.CollectionID,
			Name:      db.CollectionName,
			ShortName: db.CollectionShortName,
		},
		Model: gift.Model{
			ID:             db.ModelID,
			CollectionID:   db.CollectionID,
			Name:           db.ModelName,
			ShortName:      db.ModelShortName,
			RarityPerMille: db.ModelRarity,
		},
		Backdrop: gift.Backdrop{
			ID:             db.BackdropID,
			Name:           db.BackdropName,
			ShortName:      db.BackdropShortName,
			RarityPerMille: db.BackdropRarity,
			CenterColor:    &db.BackdropCenterColor.String,
			EdgeColor:      &db.BackdropEdgeColor.String,
			PatternColor:   &db.BackdropPatternColor.String,
			TextColor:      &db.BackdropTextColor.String,
		},
		Symbol: gift.Symbol{
			ID:             db.SymbolID,
			Name:           db.SymbolName,
			ShortName:      db.SymbolShortName,
			RarityPerMille: db.SymbolRarity,
		},
	}
}

// GiftToDomainFromGiftsByIDsRow преобразует PG модель из GetGiftsByIDsRow в domain модель
func GiftToDomainFromGiftsByIDsRow(dbGift sqlc.GetGiftsByIDsRow) *gift.Gift {
	var withdrawnAt *time.Time
	if dbGift.WithdrawnAt.Valid {
		withdrawnAt = &dbGift.WithdrawnAt.Time
	}

	return &gift.Gift{
		ID:               dbGift.ID.String(),
		OwnerTelegramID:  dbGift.OwnerTelegramID,
		Status:           gift.Status(dbGift.Status),
		Price:            dbGift.Price,
		WithdrawnAt:      withdrawnAt,
		Title:            dbGift.Title,
		Slug:             dbGift.Slug,
		CollectibleID:    dbGift.CollectibleID,
		UpgradeMessageID: dbGift.UpgradeMessageID,
		TelegramGiftID:   dbGift.TelegramGiftID,
		CreatedAt:        dbGift.CreatedAt.Time,
		UpdatedAt:        dbGift.UpdatedAt.Time,
		Collection: gift.Collection{
			ID:        dbGift.CollectionID,
			Name:      dbGift.CollectionName,
			ShortName: dbGift.CollectionShortName,
		},
		Model: gift.Model{
			ID:             dbGift.ModelID,
			CollectionID:   dbGift.CollectionID,
			Name:           dbGift.ModelName,
			ShortName:      dbGift.ModelShortName,
			RarityPerMille: 0, // Not included in GetGiftsByIDs query
		},
		Backdrop: gift.Backdrop{
			ID:             dbGift.BackdropID,
			Name:           dbGift.BackdropName,
			ShortName:      dbGift.BackdropShortName,
			RarityPerMille: 0, // Not included in GetGiftsByIDs query
		},
		Symbol: gift.Symbol{
			ID:             dbGift.SymbolID,
			Name:           dbGift.SymbolName,
			ShortName:      dbGift.SymbolShortName,
			RarityPerMille: 0, // Not included in GetGiftsByIDs query
		},
	}
}

// GiftEventToDomain преобразует PG модель события в domain модель
func GiftEventToDomain(dbEvent sqlc.GiftEvent) *gift.GiftEvent {
	var fromUserID *int64
	if dbEvent.FromUserID.Valid {
		fromUserID = &dbEvent.FromUserID.Int64
	}

	var toUserID *int64
	if dbEvent.ToUserID.Valid {
		toUserID = &dbEvent.ToUserID.Int64
	}

	var gameMode *string
	if dbEvent.GameMode.Valid {
		gameMode = &dbEvent.GameMode.String
	}

	var relatedGameID *string
	if dbEvent.RelatedGameID.Valid {
		relatedGameID = &dbEvent.RelatedGameID.String
	}

	var description *string
	if dbEvent.Description.Valid {
		description = &dbEvent.Description.String
	}

	return &gift.GiftEvent{
		ID:            dbEvent.ID.String(),
		GiftID:        dbEvent.GiftID.String(),
		FromUserID:    fromUserID,
		ToUserID:      toUserID,
		Action:        dbEvent.Action,
		GameMode:      gameMode,
		RelatedGameID: relatedGameID,
		Description:   description,
		Payload:       dbEvent.Payload,
		OccurredAt:    dbEvent.OccurredAt.Time,
	}
}

// CreateGiftParamsToDB преобразует domain параметры создания в PG параметры
func CreateGiftParamsToDB(params *gift.CreateGiftParams) sqlc.CreateGiftParams {
	return sqlc.CreateGiftParams{
		ID:               mustPgUUID(params.GiftID),
		TelegramGiftID:   params.TelegramGiftID,
		Title:            params.Title,
		Slug:             params.Slug,
		OwnerTelegramID:  params.OwnerTelegramID,
		UpgradeMessageID: params.UpgradeMessageID,
		Price:            params.Price,
		CollectibleID:    params.CollectibleID,
		Status:           sqlc.GiftStatus(params.Status),
		CreatedAt:        pgtype.Timestamptz{Time: params.CreatedAt, Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: params.UpdatedAt, Valid: true},
	}
}

// CreateCollectionParamsToDB преобразует domain параметры создания коллекции в PG параметры
func CreateCollectionParamsToDB(params *gift.CreateCollectionParams) sqlc.CreateCollectionParams {
	return sqlc.CreateCollectionParams{
		Name:      params.Name,
		ShortName: params.ShortName,
	}
}

// CreateModelParamsToDB преобразует domain параметры создания модели в PG параметры
func CreateModelParamsToDB(params *gift.CreateModelParams) sqlc.CreateModelParams {
	return sqlc.CreateModelParams{
		CollectionID:   params.CollectionID,
		Name:           params.Name,
		ShortName:      params.ShortName,
		RarityPerMille: params.RarityPerMille,
	}
}

// CreateBackdropParamsToDB преобразует domain параметры создания фона в PG параметры
func CreateBackdropParamsToDB(params *gift.CreateBackdropParams) sqlc.CreateBackdropParams {
	var centerColor, edgeColor, patternColor, textColor pgtype.Text

	if params.CenterColor != nil {
		centerColor = pgtype.Text{String: *params.CenterColor, Valid: true}
	}
	if params.EdgeColor != nil {
		edgeColor = pgtype.Text{String: *params.EdgeColor, Valid: true}
	}
	if params.PatternColor != nil {
		patternColor = pgtype.Text{String: *params.PatternColor, Valid: true}
	}
	if params.TextColor != nil {
		textColor = pgtype.Text{String: *params.TextColor, Valid: true}
	}

	return sqlc.CreateBackdropParams{
		Name:           params.Name,
		ShortName:      params.ShortName,
		RarityPerMille: params.RarityPerMille,
		CenterColor:    centerColor,
		EdgeColor:      edgeColor,
		PatternColor:   patternColor,
		TextColor:      textColor,
	}
}

// CreateSymbolParamsToDB преобразует domain параметры создания символа в PG параметры
func CreateSymbolParamsToDB(params *gift.CreateSymbolParams) sqlc.CreateSymbolParams {
	return sqlc.CreateSymbolParams{
		Name:           params.Name,
		ShortName:      params.ShortName,
		RarityPerMille: params.RarityPerMille,
	}
}

// CreateGiftEventParamsToDB преобразует domain параметры события в PG параметры
func CreateGiftEventParamsToDB(params gift.CreateGiftEventParams) sqlc.CreateGiftEventParams {
	var fromUserID pgtype.Int8
	if params.FromUserID != nil {
		fromUserID = pgtype.Int8{Int64: *params.FromUserID, Valid: true}
	}

	var toUserID pgtype.Int8
	if params.ToUserID != nil {
		toUserID = pgtype.Int8{Int64: *params.ToUserID, Valid: true}
	}

	var gameMode pgtype.Text
	if params.GameMode != nil {
		gameMode = pgtype.Text{String: *params.GameMode, Valid: true}
	}

	var relatedGameID pgtype.Text
	if params.RelatedGameID != nil {
		relatedGameID = pgtype.Text{String: *params.RelatedGameID, Valid: true}
	}

	var description pgtype.Text
	if params.Description != nil {
		description = pgtype.Text{String: *params.Description, Valid: true}
	}

	return sqlc.CreateGiftEventParams{
		GiftID:        mustPgUUID(params.GiftID),
		FromUserID:    fromUserID,
		ToUserID:      toUserID,
		RelatedGameID: relatedGameID,
		GameMode:      gameMode,
		Description:   description,
		Payload:       params.Payload,
	}
}

// ModelToDomain преобразует PG модель в domain модель
func ModelToDomain(dbModel sqlc.GiftModel) *gift.Model {
	return &gift.Model{
		ID:             dbModel.ID,
		CollectionID:   dbModel.CollectionID,
		Name:           dbModel.Name,
		ShortName:      dbModel.ShortName,
		RarityPerMille: dbModel.RarityPerMille,
	}
}

// BackdropToDomain преобразует PG модель в domain модель
func BackdropToDomain(dbBackdrop sqlc.GiftBackdrop) *gift.Backdrop {
	return &gift.Backdrop{
		ID:             dbBackdrop.ID,
		Name:           dbBackdrop.Name,
		ShortName:      dbBackdrop.ShortName,
		RarityPerMille: dbBackdrop.RarityPerMille,
		CenterColor:    stringPtrFromPgText(dbBackdrop.CenterColor),
		EdgeColor:      stringPtrFromPgText(dbBackdrop.EdgeColor),
		PatternColor:   stringPtrFromPgText(dbBackdrop.PatternColor),
		TextColor:      stringPtrFromPgText(dbBackdrop.TextColor),
	}
}

// SymbolToDomain преобразует PG модель в domain модель
func SymbolToDomain(dbSymbol sqlc.GiftSymbol) *gift.Symbol {
	return &gift.Symbol{
		ID:             dbSymbol.ID,
		Name:           dbSymbol.Name,
		ShortName:      dbSymbol.ShortName,
		RarityPerMille: dbSymbol.RarityPerMille,
	}
}

// Helper function to convert pgtype.Text to *string
func stringPtrFromPgText(pgText pgtype.Text) *string {
	if pgText.Valid {
		return &pgText.String
	}
	return nil
}
