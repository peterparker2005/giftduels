package command

import (
	"context"
	"fmt"

	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/gift/v1"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

type GiftEventHandler struct {
	repo giftDomain.Repository
	log  *zap.Logger
}

func NewGiftEventHandler(
	repo giftDomain.Repository,
	log *zap.Logger,
) *GiftEventHandler {
	return &GiftEventHandler{
		repo: repo,
		log:  log,
	}
}

// ProcessAttributesFromEvent extracts attributes from the event and creates or finds them in the database.
func (h *GiftEventHandler) ProcessAttributesFromEvent(
	ctx context.Context,
	ev *giftv1.TelegramGiftReceivedEvent,
) (*giftDomain.AttributeData, error) {
	// Process collection
	collection, err := h.repo.FindCollectionByName(ctx, ev.GetTitle())
	if err != nil {
		// Create default collection if not found
		collection, err = h.repo.CreateCollection(ctx, &giftDomain.CreateCollectionParams{
			Name:      ev.GetTitle(),
			ShortName: giftDomain.ShortName(ev.GetTitle()),
		})
		if err != nil {
			return nil, fmt.Errorf("create default collection: %w", err)
		}
	}
	collectionID := collection.ID

	// Process model
	model, err := h.repo.FindModelByName(ctx, ev.GetModel().GetName())
	if err != nil {
		// Create default model if not found
		model, err = h.repo.CreateModel(ctx, &giftDomain.CreateModelParams{
			CollectionID:   collectionID,
			Name:           ev.GetModel().GetName(),
			ShortName:      giftDomain.ShortName(ev.GetModel().GetName()),
			RarityPerMille: ev.GetModel().GetRarityPerMille(),
		})
		if err != nil {
			return nil, fmt.Errorf("create default model: %w", err)
		}
	}
	modelID := model.ID

	// Process backdrop
	backdrop, err := h.repo.FindBackdropByName(ctx, ev.GetBackdrop().GetName())
	if err != nil {
		// Create default backdrop if not found
		backdrop, err = h.repo.CreateBackdrop(ctx, &giftDomain.CreateBackdropParams{
			Name:           ev.GetBackdrop().GetName(),
			ShortName:      giftDomain.ShortName(ev.GetBackdrop().GetName()),
			RarityPerMille: ev.GetBackdrop().GetRarityPerMille(),
			CenterColor:    giftDomain.StringPtr(ev.GetBackdrop().GetCenterColor()),
			EdgeColor:      giftDomain.StringPtr(ev.GetBackdrop().GetEdgeColor()),
			PatternColor:   giftDomain.StringPtr(ev.GetBackdrop().GetPatternColor()),
			TextColor:      giftDomain.StringPtr(ev.GetBackdrop().GetTextColor()),
		})
		if err != nil {
			return nil, fmt.Errorf("create default backdrop: %w", err)
		}
	}
	backdropID := backdrop.ID

	// Process symbol
	symbol, err := h.repo.FindSymbolByName(ctx, ev.GetSymbol().GetName())
	if err != nil {
		// Create default symbol if not found
		symbol, err = h.repo.CreateSymbol(ctx, &giftDomain.CreateSymbolParams{
			Name:           ev.GetSymbol().GetName(),
			ShortName:      giftDomain.ShortName(ev.GetSymbol().GetName()),
			RarityPerMille: ev.GetSymbol().GetRarityPerMille(),
		})
		if err != nil {
			return nil, fmt.Errorf("create default symbol: %w", err)
		}
	}
	symbolID := symbol.ID

	return &giftDomain.AttributeData{
		CollectionID: collectionID,
		ModelID:      modelID,
		BackdropID:   backdropID,
		SymbolID:     symbolID,
	}, nil
}

// CreateGiftFromEvent creates a gift from event data.
func (h *GiftEventHandler) CreateGiftFromEvent(
	ctx context.Context,
	giftID string,
	ev *giftv1.TelegramGiftReceivedEvent,
	price *tonamount.TonAmount,
	attrs *giftDomain.AttributeData,
) (*giftDomain.Gift, error) {
	createParams := &giftDomain.CreateGiftParams{
		GiftID:           giftID,
		OwnerTelegramID:  ev.GetOwnerTelegramId().GetValue(),
		Price:            price,
		Title:            ev.GetTitle(),
		Slug:             ev.GetSlug(),
		CollectibleID:    ev.GetCollectibleId(),
		UpgradeMessageID: ev.GetUpgradeMessageId(),
		TelegramGiftID:   ev.GetTelegramGiftId().GetValue(),
	}

	return h.repo.CreateGift(
		ctx,
		createParams,
		attrs.CollectionID,
		attrs.ModelID,
		attrs.BackdropID,
		attrs.SymbolID,
	)
}
