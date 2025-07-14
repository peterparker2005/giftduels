package gift

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
)

type CreateGiftParams struct {
	GiftID           string
	OwnerTelegramID  int64
	CollectibleID    int32
	UpgradeMessageID int32
	TelegramGiftID   int64
	Status           Status
	Price            *tonamount.TonAmount
	Title            string
	Slug             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateCollectionParams struct {
	Name      string
	ShortName string
}

type CreateModelParams struct {
	CollectionID   int32
	Name           string
	ShortName      string
	RarityPerMille int32
}

type CreateBackdropParams struct {
	Name           string
	ShortName      string
	RarityPerMille int32
	CenterColor    *string
	EdgeColor      *string
	PatternColor   *string
	TextColor      *string
}

type CreateSymbolParams struct {
	Name           string
	ShortName      string
	RarityPerMille int32
}

type GetUserGiftsResult struct {
	Gifts []*Gift
	Total int64
}

type CreateGiftEventParams struct {
	GiftID         string
	TelegramUserID int64
	EventType      EventType
	RelatedGameID  *string
}

type Repository interface {
	WithTx(tx pgx.Tx) Repository
	GetGiftByID(ctx context.Context, id string) (*Gift, error)
	GetUserGifts(
		ctx context.Context,
		limit int32,
		offset int32,
		ownerTelegramID int64,
	) (*GetUserGiftsResult, error)
	GetUserActiveGifts(
		ctx context.Context,
		limit int32,
		offset int32,
		ownerTelegramID int64,
	) (*GetUserGiftsResult, error)
	StakeGiftForGame(ctx context.Context, id string) (*Gift, error)
	ReturnGiftFromGame(ctx context.Context, id string) (*Gift, error)
	UpdateGiftOwner(ctx context.Context, id string, ownerTelegramID int64) (*Gift, error)
	MarkGiftForWithdrawal(ctx context.Context, id string) (*Gift, error)
	CancelGiftWithdrawal(ctx context.Context, id string) (*Gift, error)
	CompleteGiftWithdrawal(ctx context.Context, id string) (*Gift, error)
	CreateGift(
		ctx context.Context,
		params *CreateGiftParams,
		collectionID, modelID, backdropID, symbolID int32,
	) (*Gift, error)
	CreateGiftEvent(ctx context.Context, params CreateGiftEventParams) (*Event, error)
	GetGiftEvents(
		ctx context.Context,
		giftID string,
		limit int32,
		offset int32,
	) ([]*Event, error)
	GetGiftsByIDs(ctx context.Context, ids []string) ([]*Gift, error)
	SaveGiftWithPrice(ctx context.Context, id string, price *tonamount.TonAmount) (*Gift, error)

	// Lookup table methods
	GetGiftModel(ctx context.Context, id int32) (*Model, error)
	GetGiftBackdrop(ctx context.Context, id int32) (*Backdrop, error)
	GetGiftSymbol(ctx context.Context, id int32) (*Symbol, error)
	GetGiftCollection(ctx context.Context, id int32) (*Collection, error)

	// Create lookup table methods
	CreateCollection(ctx context.Context, params *CreateCollectionParams) (*Collection, error)
	CreateModel(ctx context.Context, params *CreateModelParams) (*Model, error)
	CreateBackdrop(ctx context.Context, params *CreateBackdropParams) (*Backdrop, error)
	CreateSymbol(ctx context.Context, params *CreateSymbolParams) (*Symbol, error)

	// Find lookup table methods
	FindCollectionByName(ctx context.Context, name string) (*Collection, error)
	FindModelByName(ctx context.Context, name string) (*Model, error)
	FindBackdropByName(ctx context.Context, name string) (*Backdrop, error)
	FindSymbolByName(ctx context.Context, name string) (*Symbol, error)
}
