package pricing

import "context"

type PriceServiceParams struct {
	Collection string
	Model      string
	Backdrop   string
	Symbol     string
}

type PriceServiceResult struct {
	FloorPrice string
	Price      string
}

type PriceService interface {
	GetFloorPrice(ctx context.Context, params *PriceServiceParams) (*PriceServiceResult, error)
}
