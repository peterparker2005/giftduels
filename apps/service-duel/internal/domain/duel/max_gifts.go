package duel

type MaxGifts uint32

const (
	MinGifts      = 1
	MaxGiftsLimit = 10
)

func NewMaxGifts(n uint32) (MaxGifts, error) {
	if n < MinGifts || n > MaxGiftsLimit {
		return 0, ErrInvalidMaxGifts
	}
	return MaxGifts(n), nil
}

func (m MaxGifts) Uint32() uint32 {
	return uint32(m)
}

func (m MaxGifts) Int32() int32 {
	//nolint:gosec // we know that m is in range
	return int32(m)
}
