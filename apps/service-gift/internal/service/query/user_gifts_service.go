package query

import (
	"context"

	"github.com/ccoveille/go-safecast"
	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"github.com/peterparker2005/giftduels/packages/shared"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

type UserGiftsService struct {
	repo              giftDomain.Repository
	log               *logger.Logger
	duelPrivateClient duelv1.DuelPrivateServiceClient
}

func NewUserGiftsService(
	repo giftDomain.Repository,
	log *logger.Logger,
	clients *clients.Clients,
) *UserGiftsService {
	return &UserGiftsService{
		repo:              repo,
		log:               log,
		duelPrivateClient: clients.Duel.Private,
	}
}

type GetUserGiftsResult struct {
	Gifts      []*giftDomain.Gift
	Total      int32
	TotalValue *tonamount.TonAmount
}

func (s *UserGiftsService) GetUserGifts(
	ctx context.Context,
	telegramUserID int64,
	pagination *shared.PageRequest,
) (*GetUserGiftsResult, error) {
	res, err := s.repo.GetUserGifts(ctx, pagination.PageSize(), pagination.Offset(), telegramUserID)
	if err != nil {
		return nil, err
	}

	// Populate attributes for all gifts
	if err = s.populateGiftAttributes(ctx, res.Gifts); err != nil {
		return nil, err
	}

	totalValue, err := tonamount.NewTonAmountFromNano(0) // Start with zero TON amount
	if err != nil {
		return nil, err
	}
	for _, g := range res.Gifts {
		if g.Price != nil {
			totalValue = totalValue.Add(g.Price)
		}
	}

	total, err := safecast.ToInt32(res.Total)
	if err != nil {
		return nil, err
	}

	for _, g := range res.Gifts {
		duelID, err := s.findDuelByGiftID(ctx, g.ID)
		if err != nil {
			return nil, err
		}
		g.SetRelatedDuelID(duelID)
	}

	return &GetUserGiftsResult{
		Gifts:      res.Gifts,
		Total:      total,
		TotalValue: totalValue,
	}, nil
}

func (s *UserGiftsService) GetUserActiveGifts(
	ctx context.Context,
	telegramUserID int64,
	pagination *shared.PageRequest,
) (*GetUserGiftsResult, error) {
	res, err := s.repo.GetUserActiveGifts(
		ctx,
		pagination.PageSize(),
		pagination.Offset(),
		telegramUserID,
	)
	if err != nil {
		s.log.Error("Failed to get user active gifts", zap.Error(err))
		return nil, err
	}

	// Populate attributes for all gifts
	if err = s.populateGiftAttributes(ctx, res.Gifts); err != nil {
		s.log.Error("Failed to populate gift attributes", zap.Error(err))
		return nil, err
	}

	// TODO: calculate total value in one query
	totalValue, err := tonamount.NewTonAmountFromNano(0) // Start with zero TON amount
	if err != nil {
		s.log.Error("Failed to calculate total value", zap.Error(err))
		return nil, err
	}
	for _, g := range res.Gifts {
		if g.Price != nil {
			totalValue = totalValue.Add(g.Price)
		}
	}

	total, err := safecast.ToInt32(res.Total)
	if err != nil {
		s.log.Error("Failed to calculate total", zap.Error(err))
		return nil, err
	}

	for _, g := range res.Gifts {
		if g.Status == giftDomain.StatusInGame {
			duelID, err := s.findDuelByGiftID(ctx, g.ID)
			if err != nil {
				s.log.Error("Failed to find duel by gift ID", zap.Error(err))
			}
			if duelID == "" {
				s.log.Error("Duel not found for gift", zap.String("giftID", g.ID))
				continue
			}
			g.SetRelatedDuelID(duelID)
		}
	}

	return &GetUserGiftsResult{
		Gifts:      res.Gifts,
		Total:      total,
		TotalValue: totalValue,
	}, nil
}

func (s *UserGiftsService) findDuelByGiftID(ctx context.Context, giftID string) (string, error) {
	duelID, err := s.duelPrivateClient.FindDuelByGiftID(ctx, &duelv1.FindDuelByGiftIDRequest{
		GiftId: &sharedv1.GiftId{Value: giftID},
	})
	if err != nil {
		return "", err
	}
	return duelID.GetDuelId().GetValue(), nil
}

// populateGiftAttributes populates the Model, Backdrop, and Symbol attributes for a slice of gifts.
func (s *UserGiftsService) populateGiftAttributes(
	ctx context.Context,
	gifts []*giftDomain.Gift,
) error {
	// Collect all unique IDs
	modelIDs, backdropIDs, symbolIDs := s.collectAttributeIDs(gifts)

	// Fetch all models, backdrops, and symbols in parallel
	models, backdrops, symbols, err := s.fetchAttributesInParallel(
		ctx,
		modelIDs,
		backdropIDs,
		symbolIDs,
	)
	if err != nil {
		return err
	}

	// Populate the gifts with the fetched data
	s.populateGiftsWithAttributes(gifts, models, backdrops, symbols)

	return nil
}

func (s *UserGiftsService) collectAttributeIDs(
	gifts []*giftDomain.Gift,
) (map[int32]bool, map[int32]bool, map[int32]bool) {
	modelIDs := make(map[int32]bool)
	backdropIDs := make(map[int32]bool)
	symbolIDs := make(map[int32]bool)

	for _, gift := range gifts {
		modelIDs[gift.Model.ID] = true
		backdropIDs[gift.Backdrop.ID] = true
		symbolIDs[gift.Symbol.ID] = true
	}

	return modelIDs, backdropIDs, symbolIDs
}

func (s *UserGiftsService) fetchAttributesInParallel(
	ctx context.Context,
	modelIDs, backdropIDs, symbolIDs map[int32]bool,
) (map[int32]*giftDomain.Model, map[int32]*giftDomain.Backdrop, map[int32]*giftDomain.Symbol, error) {
	//nolint:mnd // 3 is not a magic number
	errorChan := make(chan error, 3)
	modelChan := make(chan map[int32]*giftDomain.Model, 1)
	backdropChan := make(chan map[int32]*giftDomain.Backdrop, 1)
	symbolChan := make(chan map[int32]*giftDomain.Symbol, 1)

	// Fetch models
	go s.fetchModels(ctx, modelIDs, modelChan, errorChan)

	// Fetch backdrops
	go s.fetchBackdrops(ctx, backdropIDs, backdropChan, errorChan)

	// Fetch symbols
	go s.fetchSymbols(ctx, symbolIDs, symbolChan, errorChan)

	// Wait for all goroutines to complete
	models := <-modelChan
	backdrops := <-backdropChan
	symbols := <-symbolChan

	// Check for errors
	select {
	case err := <-errorChan:
		return nil, nil, nil, err
	default:
	}

	return models, backdrops, symbols, nil
}

func (s *UserGiftsService) fetchModels(
	ctx context.Context,
	modelIDs map[int32]bool,
	modelChan chan<- map[int32]*giftDomain.Model,
	errorChan chan<- error,
) {
	models := make(map[int32]*giftDomain.Model)
	for modelID := range modelIDs {
		model, err := s.repo.GetGiftModel(ctx, modelID)
		if err != nil {
			errorChan <- err
			return
		}
		models[modelID] = model
	}
	modelChan <- models
}

func (s *UserGiftsService) fetchBackdrops(
	ctx context.Context,
	backdropIDs map[int32]bool,
	backdropChan chan<- map[int32]*giftDomain.Backdrop,
	errorChan chan<- error,
) {
	backdrops := make(map[int32]*giftDomain.Backdrop)
	for backdropID := range backdropIDs {
		backdrop, err := s.repo.GetGiftBackdrop(ctx, backdropID)
		if err != nil {
			errorChan <- err
			return
		}
		backdrops[backdropID] = backdrop
	}
	backdropChan <- backdrops
}

func (s *UserGiftsService) fetchSymbols(
	ctx context.Context,
	symbolIDs map[int32]bool,
	symbolChan chan<- map[int32]*giftDomain.Symbol,
	errorChan chan<- error,
) {
	symbols := make(map[int32]*giftDomain.Symbol)
	for symbolID := range symbolIDs {
		symbol, err := s.repo.GetGiftSymbol(ctx, symbolID)
		if err != nil {
			errorChan <- err
			return
		}
		symbols[symbolID] = symbol
	}
	symbolChan <- symbols
}

func (s *UserGiftsService) populateGiftsWithAttributes(
	gifts []*giftDomain.Gift,
	models map[int32]*giftDomain.Model,
	backdrops map[int32]*giftDomain.Backdrop,
	symbols map[int32]*giftDomain.Symbol,
) {
	for _, gift := range gifts {
		s.populateSingleGiftAttributes(gift, models, backdrops, symbols)
	}
}

func (s *UserGiftsService) populateSingleGiftAttributes(
	gift *giftDomain.Gift,
	models map[int32]*giftDomain.Model,
	backdrops map[int32]*giftDomain.Backdrop,
	symbols map[int32]*giftDomain.Symbol,
) {
	if model, exists := models[gift.Model.ID]; exists && model != nil {
		gift.Model = *model
	} else {
		s.log.Warn("Model not found for gift", zap.String("giftID", gift.ID), zap.Int32("modelID", gift.Model.ID))
	}
	if backdrop, exists := backdrops[gift.Backdrop.ID]; exists && backdrop != nil {
		gift.Backdrop = *backdrop
	} else {
		s.log.Warn("Backdrop not found for gift", zap.String("giftID", gift.ID), zap.Int32("backdropID", gift.Backdrop.ID))
	}
	if symbol, exists := symbols[gift.Symbol.ID]; exists && symbol != nil {
		gift.Symbol = *symbol
	} else {
		s.log.Warn("Symbol not found for gift", zap.String("giftID", gift.ID), zap.Int32("symbolID", gift.Symbol.ID))
	}
}
