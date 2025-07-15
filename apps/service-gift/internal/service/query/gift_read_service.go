package query

import (
	"context"

	giftDomain "github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/packages/grpc-go/clients"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	sharedv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/shared/v1"
	"go.uber.org/zap"
)

type GiftReadService struct {
	repo              giftDomain.Repository
	log               *logger.Logger
	duelPrivateClient duelv1.DuelPrivateServiceClient
}

func NewGiftReadService(
	repo giftDomain.Repository,
	log *logger.Logger,
	clients *clients.Clients,
) *GiftReadService {
	return &GiftReadService{
		repo:              repo,
		log:               log,
		duelPrivateClient: clients.Duel.Private,
	}
}

func (s *GiftReadService) GetGiftByID(ctx context.Context, id string) (*giftDomain.Gift, error) {
	gift, err := s.repo.GetGiftByID(ctx, id)
	if err != nil {
		return nil, err
	}

	collection, err := s.repo.FindCollectionByName(
		ctx,
		gift.Title,
	) // Using title as collection name
	if err != nil {
		// If collection not found, create a default one
		collection, err = s.repo.CreateCollection(ctx, &giftDomain.CreateCollectionParams{
			Name:      gift.Title,
			ShortName: giftDomain.ShortName(gift.Title),
		})
		if err != nil {
			return nil, err
		}
	}

	model, err := s.repo.GetGiftModel(ctx, gift.Model.ID)
	if err != nil {
		return nil, err
	}
	backdrop, err := s.repo.GetGiftBackdrop(ctx, gift.Backdrop.ID)
	if err != nil {
		return nil, err
	}
	symbol, err := s.repo.GetGiftSymbol(ctx, gift.Symbol.ID)
	if err != nil {
		return nil, err
	}

	// Add nil checks to prevent panic
	if collection != nil {
		gift.Collection = *collection
	}
	if model != nil {
		gift.Model = *model
	}
	if backdrop != nil {
		gift.Backdrop = *backdrop
	}
	if symbol != nil {
		gift.Symbol = *symbol
	}

	if gift.Status == giftDomain.StatusInGame {
		duelID, err := s.findDuelByGiftID(ctx, gift.ID)
		if err != nil {
			return nil, err
		}
		gift.SetRelatedDuelID(duelID)
	}

	return gift, nil
}

func (s *GiftReadService) GetGiftsByIDs(
	ctx context.Context,
	giftIDs []string,
) ([]*giftDomain.Gift, error) {
	gifts, err := s.repo.GetGiftsByIDs(ctx, giftIDs)
	if err != nil {
		return nil, err
	}

	// Populate attributes for all gifts
	if err = s.populateGiftAttributes(ctx, gifts); err != nil {
		return nil, err
	}

	return gifts, nil
}

func (s *GiftReadService) findDuelByGiftID(ctx context.Context, giftID string) (string, error) {
	duelID, err := s.duelPrivateClient.FindDuelByGiftID(ctx, &duelv1.FindDuelByGiftIDRequest{
		GiftId: &sharedv1.GiftId{Value: giftID},
	})
	if err != nil {
		return "", err
	}
	return duelID.GetDuelId().GetValue(), nil
}

// populateGiftAttributes populates the Model, Backdrop, and Symbol attributes for a slice of gifts.
func (s *GiftReadService) populateGiftAttributes(
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

func (s *GiftReadService) collectAttributeIDs(
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

func (s *GiftReadService) fetchAttributesInParallel(
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

func (s *GiftReadService) fetchModels(
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

func (s *GiftReadService) fetchBackdrops(
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

func (s *GiftReadService) fetchSymbols(
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

func (s *GiftReadService) populateGiftsWithAttributes(
	gifts []*giftDomain.Gift,
	models map[int32]*giftDomain.Model,
	backdrops map[int32]*giftDomain.Backdrop,
	symbols map[int32]*giftDomain.Symbol,
) {
	for _, gift := range gifts {
		s.populateSingleGiftAttributes(gift, models, backdrops, symbols)
	}
}

func (s *GiftReadService) populateSingleGiftAttributes(
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
