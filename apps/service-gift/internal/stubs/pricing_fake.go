package stubs

import (
	"context"
	"math/rand"
	"time"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/pricing"
)

type fakeRepo struct {
	names []string // пул вымышленных названий
	min   float64  // нижняя граница TON-цены
	max   float64  // верхняя граница TON-цены
}

// NewPricingFake возвращает реализацию pricing.Repository,
// генерирующую limit случайных Observation.
func NewPricingFake() pricing.Repository {
	rand.Seed(time.Now().UnixNano())
	return &fakeRepo{
		names: []string{
			"Neon Rat", "Wolf Rage", "Spiced Wine",
			"Golden Hamster", "Purple Tiger", "Crystal Onion",
		},
		min: 0.1, // 0.1 TON
		max: 5.0, // 5 TON
	}
}

func (r *fakeRepo) Samples(
	ctx context.Context,
	_ pricing.Filter, // фильтр игнорируем — это фейк
	limit int,
) ([]pricing.Observation, error) {
	if limit <= 0 {
		limit = 1
	}
	out := make([]pricing.Observation, limit)

	now := time.Now().Unix()
	for i := 0; i < limit; i++ {
		out[i] = pricing.Observation{
			GiftName:  r.names[rand.Intn(len(r.names))],
			TonPrice:  r.min + rand.Float64()*(r.max-r.min),
			Timestamp: now,
			Source:    "fake",
		}
	}
	return out, nil
}
