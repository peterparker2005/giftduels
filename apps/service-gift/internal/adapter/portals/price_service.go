package portals

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/pricing"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"github.com/peterparker2005/giftduels/packages/tonamount-go"
	"go.uber.org/zap"
)

type portalsPriceService struct {
	client *HTTPClient
	logger *logger.Logger
}

func NewPortalsPriceService(logger *logger.Logger) pricing.PriceService {
	client := NewHTTPClient("", logger)
	return &portalsPriceService{client: client, logger: logger}
}

func (s *portalsPriceService) GetFloorPrice(
	ctx context.Context,
	params *pricing.PriceServiceParams,
) (*pricing.PriceServiceResult, error) {
	if params == nil {
		s.logger.Error("params is nil")
		return nil, errors.New("params is nil")
	}

	type attempt struct {
		model    string
		symbol   string
		backdrop string
	}

	// Определяем fallback порядок: убираем по одному параметру справа налево
	attempts := []attempt{
		{params.Model, params.Symbol, params.Backdrop}, // full
		{params.Model, "", params.Backdrop},            // no symbol
		{params.Model, "", ""},                         // no symbol, no backdrop
		{"", "", ""},                                   // only collection
	}

	var lastErr error
	for i, a := range attempts {
		s.logger.Info(
			"Try fetching floor price",
			zap.Int("attempt", i),
			zap.String("collection", params.Collection),
			zap.String("model", a.model),
			zap.String("symbol", a.symbol),
			zap.String("backdrop", a.backdrop),
		)

		resp, err := retryWithBackoff(func() (*NFTResponse, error) {
			return s.client.SearchNFTs(ctx, params.Collection, a.model, a.symbol, a.backdrop)
		})
		if err != nil {
			s.logger.Warn("Fallback search failed", zap.Int("attempt", i), zap.Error(err))
			lastErr = err
			continue
		}

		// При успехе парсим
		priceStr := resp.Results[0].FloorPrice
		if priceStr == "" {
			priceStr = resp.Results[0].Price
		}
		price, err := tonamount.NewTonAmountFromString(priceStr)
		if err != nil {
			s.logger.Error("parse price failed", zap.String("price", priceStr), zap.Error(err))
			return nil, fmt.Errorf("parse price: %w", err)
		}

		return &pricing.PriceServiceResult{
			FloorPrice: price.String(),
			Price:      price.String(),
		}, nil
	}

	return nil, fmt.Errorf("all attempts failed, last error: %w", lastErr)
}

func retryWithBackoff(fn func() (*NFTResponse, error)) (*NFTResponse, error) {
	backoff := defaultBackoffDelay
	for i := range 3 {
		resp, err := fn()
		if err != nil {
			if isTooManyRequests(err) && i < 2 {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, err
		}
		return resp, nil
	}
	return nil, errors.New("too many retries")
}

func isTooManyRequests(err error) bool {
	return strings.Contains(err.Error(), "429")
}
