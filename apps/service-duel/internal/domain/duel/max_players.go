package duel

type MaxPlayers uint32

const (
	MinPlayers      = 2
	MaxPlayersLimit = 4
)

func NewMaxPlayers(n uint32) (MaxPlayers, error) {
	if n < MinPlayers || n > MaxPlayersLimit {
		return 0, ErrInvalidMaxPlayers
	}
	return MaxPlayers(n), nil
}

func (m MaxPlayers) Uint32() uint32 {
	return uint32(m)
}

func (m MaxPlayers) Int32() int32 {
	//nolint:gosec // we know that m is in range
	return int32(m)
}
