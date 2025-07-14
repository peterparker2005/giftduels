package gift

import "context"

// BulkLoader can load different types of lookup entities at the same time.
type BulkLoader struct {
	// functions to access the repository
	GetModels    func(ctx context.Context, ids []int32) ([]*Model, error)
	GetBackdrops func(ctx context.Context, ids []int32) ([]*Backdrop, error)
	GetSymbols   func(ctx context.Context, ids []int32) ([]*Symbol, error)
}

// LoadAttributes collects models, backdrops and symbols in 3 requests,
// and returns maps id→entity for each of them.
func (b *BulkLoader) LoadAttributes(
	ctx context.Context,
	gifts []*Gift, // your domain model Gift, with fields Model.ID, Backdrop.ID, Symbol.ID
) (
	map[int32]*Model,
	map[int32]*Backdrop,
	map[int32]*Symbol,
	error,
) {
	// 1) Collect unique IDs
	modelIDs := uniqueInts32(collect(gifts, func(g *Gift) int32 { return g.Model.ID }))
	backdropIDs := uniqueInts32(collect(gifts, func(g *Gift) int32 { return g.Backdrop.ID }))
	symbolIDs := uniqueInts32(collect(gifts, func(g *Gift) int32 { return g.Symbol.ID }))

	var err error

	// 2) Make 3 batch requests
	var (
		modelList    []*Model
		backdropList []*Backdrop
		symbolList   []*Symbol
	)
	if modelList, err = b.GetModels(ctx, modelIDs); err != nil {
		return nil, nil, nil, err
	}
	if backdropList, err = b.GetBackdrops(ctx, backdropIDs); err != nil {
		return nil, nil, nil, err
	}
	if symbolList, err = b.GetSymbols(ctx, symbolIDs); err != nil {
		return nil, nil, nil, err
	}
	models := make(map[int32]*Model, len(modelList))
	for _, m := range modelList {
		models[m.ID] = m
	}

	backdrops := make(map[int32]*Backdrop, len(backdropList))
	for _, d := range backdropList {
		backdrops[d.ID] = d
	}

	symbols := make(map[int32]*Symbol, len(symbolList))
	for _, s := range symbolList {
		symbols[s.ID] = s
	}

	return models, backdrops, symbols, err
}

// Collect collects values from a slice of gifts using a function.
func collect[T any](gifts []*Gift, fn func(*Gift) T) []T {
	out := make([]T, 0, len(gifts))
	for _, g := range gifts {
		out = append(out, fn(g))
	}
	return out
}

// UniqueInts32 returns a slice of unique int32 values from a slice of int32 values.
func uniqueInts32(src []int32) []int32 {
	seen := map[int32]struct{}{}
	out := make([]int32, 0, len(src))
	for _, v := range src {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
