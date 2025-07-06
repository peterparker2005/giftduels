package pricing

import "context"

// Repository — то, что умеют реализовать Tonnel, Portals, GiftsAPI, Fake…
type Repository interface {
	// Samples возвращает ≤limit наблюдений по подаркам,
	// подходящим под фильтр.  Если найдено <limit,     —
	// это ОК: репо отдает всё, что смог.
	Samples(ctx context.Context, f Filter, limit int) ([]Observation, error)
}
